package main

import (
    "encoding/json"
    "fmt"
    "net/http"
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


