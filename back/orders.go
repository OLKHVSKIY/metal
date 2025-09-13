package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// Order represents a service order submitted from services page
type Order struct {
    ID        int64     `json:"id"`
    Service   string    `json:"service"`
    Name      string    `json:"name"`
    Phone     string    `json:"phone"`
    Email     string    `json:"email"`
    Status    string    `json:"status"` // active/closed
    CreatedAt time.Time `json:"createdAt"`
}

func createOrder(o *Order) (int64, error) {
    res, err := db.Exec("INSERT INTO orders (service, name, phone, email, status) VALUES (?,?,?,?,?)", o.Service, o.Name, o.Phone, o.Email, o.Status)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func listOrders() ([]Order, error) {
    rows, err := db.Query("SELECT id, service, name, phone, email, status, created_at FROM orders ORDER BY created_at DESC")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Order
    for rows.Next() {
        var o Order
        var ts string
        if err := rows.Scan(&o.ID, &o.Service, &o.Name, &o.Phone, &o.Email, &o.Status, &ts); err != nil {
            return nil, err
        }
        t, _ := time.Parse("2006-01-02 15:04:05", ts)
        o.CreatedAt = t
        out = append(out, o)
    }
    return out, nil
}

func updateOrderStatus(id int64, status string) error {
    _, err := db.Exec("UPDATE orders SET status=? WHERE id=?", status, id)
    return err
}

// --- Quick item orders (from catalog item page) ---
type ItemOrder struct {
    ID        int64   `json:"id"`
    ItemID    string  `json:"item_id"`
    Title     string  `json:"title"`
    Qty       int     `json:"qty"`
    Price     float64 `json:"price"`
    Total     float64 `json:"total"`
    Phone     string  `json:"phone"`
    UserLogin string  `json:"user_login"`
    Status    string  `json:"status"`
    CreatedAt string  `json:"created_at"`
}

func handleCreateItemOrder(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    var p struct {
        ItemID string  `json:"item_id"`
        Title  string  `json:"title"`
        Qty    int     `json:"qty"`
        Price  float64 `json:"price"`
        Phone  string  `json:"phone"`
    }
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
    if p.Qty <= 0 { p.Qty = 1 }
    total := float64(p.Qty) * p.Price
    // read user login from cookie if present
    var login string
    if c, err := r.Cookie(userSessionCookieName); err == nil && c.Value != "" {
        login, _ = url.QueryUnescape(c.Value)
    }
    // If phone is not provided but user is authenticated, try to read phone from users table
    phone := strings.TrimSpace(p.Phone)
    if phone == "" && strings.TrimSpace(login) != "" {
        var ph sql.NullString
        if err := userDB.QueryRow("SELECT phone FROM users WHERE login=?", strings.TrimSpace(login)).Scan(&ph); err == nil {
            if ph.Valid { phone = strings.TrimSpace(ph.String) }
        }
    }
    res, err := itemOrdersDB.Exec("INSERT INTO item_orders(item_id,title,qty,price,total,phone,user_login) VALUES(?,?,?,?,?,?,?)", strings.TrimSpace(p.ItemID), strings.TrimSpace(p.Title), p.Qty, p.Price, total, phone, strings.TrimSpace(login))
    if err != nil { http.Error(w, err.Error(), 500); return }
    id, _ := res.LastInsertId()
    // Telegram notify
    msg := fmt.Sprintf("🛒 Новый заказ одним кликом\nТовар: %s\nКол-во: %d\nСумма: %.2f", p.Title, p.Qty, total)
    if strings.TrimSpace(login) != "" { msg += "\nПользователь: " + strings.TrimSpace(login) }
    if phone != "" { msg += "\nТелефон: " + phone }
    go sendTelegram(msg)
    writeJSON(w, map[string]any{"id": id, "status":"ok"})
}

func adminItemOrdersList(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    rows, err := itemOrdersDB.Query("SELECT id,item_id,title,qty,price,total,phone,user_login,status,created_at FROM item_orders ORDER BY created_at DESC")
    if err != nil { http.Error(w, err.Error(), 500); return }
    defer rows.Close()
    var out []ItemOrder
    for rows.Next() {
        var o ItemOrder
        if err := rows.Scan(&o.ID,&o.ItemID,&o.Title,&o.Qty,&o.Price,&o.Total,&o.Phone,&o.UserLogin,&o.Status,&o.CreatedAt); err == nil { out = append(out, o) }
    }
    writeJSON(w, out)
}

func adminItemOrdersStatus(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPatch { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    var p struct{ ID int64 `json:"id"`; Status string `json:"status"` }
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
    if p.ID == 0 || strings.TrimSpace(p.Status)=="" { http.Error(w, "id and status required", 400); return }
    _, err := itemOrdersDB.Exec("UPDATE item_orders SET status=? WHERE id=?", p.Status, p.ID)
    if err != nil { http.Error(w, err.Error(), 500); return }
    writeJSON(w, map[string]string{"status":"ok"})
}

// Batch: create multiple item orders and send one Telegram message
func handleCreateItemOrderBatch(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    var payload struct {
        Items []struct{
            ItemID string  `json:"item_id"`
            Title  string  `json:"title"`
            Qty    int     `json:"qty"`
            Price  float64 `json:"price"`
        } `json:"items"`
        Phone string `json:"phone"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { http.Error(w, "bad json", 400); return }
    if len(payload.Items) == 0 { http.Error(w, "empty items", 400); return }
    // login + phone enrichment
    var login string
    if c, err := r.Cookie(userSessionCookieName); err == nil && c.Value != "" { login, _ = url.QueryUnescape(c.Value) }
    phone := strings.TrimSpace(payload.Phone)
    if phone == "" && strings.TrimSpace(login) != "" {
        var ph sql.NullString
        if err := userDB.QueryRow("SELECT phone FROM users WHERE login=?", strings.TrimSpace(login)).Scan(&ph); err == nil {
            if ph.Valid { phone = strings.TrimSpace(ph.String) }
        }
    }
    // insert all
    tx, err := itemOrdersDB.Begin()
    if err != nil { http.Error(w, err.Error(), 500); return }
    for _, it := range payload.Items {
        qty := it.Qty; if qty <= 0 { qty = 1 }
        total := float64(qty) * it.Price
        if _, err := tx.Exec("INSERT INTO item_orders(item_id,title,qty,price,total,phone,user_login) VALUES(?,?,?,?,?,?,?)", strings.TrimSpace(it.ItemID), strings.TrimSpace(it.Title), qty, it.Price, total, phone, strings.TrimSpace(login)); err != nil {
            tx.Rollback(); http.Error(w, err.Error(), 500); return
        }
    }
    if err := tx.Commit(); err != nil { http.Error(w, err.Error(), 500); return }
    // Telegram single message
    var b strings.Builder
    b.WriteString("🛒 Новый заказ из корзины\n")
    if strings.TrimSpace(login) != "" { b.WriteString("Пользователь: "); b.WriteString(strings.TrimSpace(login)); b.WriteString("\n") }
    if phone != "" { b.WriteString("Телефон: "); b.WriteString(phone); b.WriteString("\n") }
    var totalSum float64
    for _, it := range payload.Items {
        qty := it.Qty; if qty <= 0 { qty = 1 }
        itemTotal := float64(qty) * it.Price
        totalSum += itemTotal
        b.WriteString("• "); b.WriteString(strings.TrimSpace(it.Title)); b.WriteString(" — "); b.WriteString(fmt.Sprintf("%d шт. — %.2f ₽", qty, itemTotal)); b.WriteString("\n")
    }
    b.WriteString("Итого: "); b.WriteString(fmt.Sprintf("%.2f ₽", totalSum))
    go sendTelegram(b.String())
    writeJSON(w, map[string]string{"status":"ok"})
}

// Public orders endpoint - accepts new orders
func ordersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        var payload struct {
            Service string `json:"service"`
            Name    string `json:"name"`
            Phone   string `json:"phone"`
            Email   string `json:"email"`
        }
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        o := &Order{
            Service: payload.Service,
            Name:    strings.TrimSpace(payload.Name),
            Phone:   strings.TrimSpace(payload.Phone),
            Email:   strings.TrimSpace(payload.Email),
            Status:  "active",
        }
        id, err := createOrder(o)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        // Telegram notify (best-effort)
        go sendTelegram(fmt.Sprintf("🛠 Новая заявка: %s\nИмя: %s\nТелефон: %s\nEmail: %s", o.Service, o.Name, o.Phone, o.Email))
        writeJSON(w, map[string]any{"id": id, "status": "ok"})
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

// Admin list and status update
func adminListOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    items, err := listOrders()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, items)
}

func adminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPatch {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var payload struct {
        ID     int64  `json:"id"`
        Status string `json:"status"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if payload.Status != "active" && payload.Status != "closed" {
        http.Error(w, "invalid status", http.StatusBadRequest)
        return
    }
    if err := updateOrderStatus(payload.ID, payload.Status); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, map[string]string{"status": "ok"})
}


