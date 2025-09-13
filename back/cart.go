package main

import (
    "encoding/json"
    "net/http"
    "strings"
)

// cartHandler persists simple carts in SQLite keyed by anonymous cart_id cookie.
// This enables cross-device continuity without requiring auth.
func cartHandler(w http.ResponseWriter, r *http.Request) {
    // Resolve cart id from cookie; create if missing
    c, err := r.Cookie("cart_id")
    cartID := ""
    if err == nil { cartID = strings.TrimSpace(c.Value) }
    if cartID == "" {
        // use csrf token or random fallback
        cartID = generateCSRFToken()
        http.SetCookie(w, &http.Cookie{ Name: "cart_id", Value: cartID, Path: "/", MaxAge: 60*60*24*90 })
    }

    switch r.Method {
    case http.MethodGet:
        rows, err := cartDB.Query("SELECT item_id, title, price, image, qty FROM cart_items WHERE cart_id=?", cartID)
        if err != nil { writeJSON(w, []any{}); return }
        defer rows.Close()
        type item struct{ ID, Title, Image string; Price float64; Qty int }
        var out []item
        for rows.Next() {
            var it item
            if err := rows.Scan(&it.ID, &it.Title, &it.Price, &it.Image, &it.Qty); err == nil { out = append(out, it) }
        }
        writeJSON(w, out)
    case http.MethodPost:
        var p struct {
            ID    string  `json:"id"`
            Title string  `json:"title"`
            Price float64 `json:"price"`
            Image string  `json:"image"`
            Qty   int     `json:"qty"`
        }
        if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
        p.ID = strings.TrimSpace(p.ID)
        if p.ID == "" { http.Error(w, "id required", 400); return }
        if p.Qty <= 0 { p.Qty = 1 }
        // upsert
        tx, _ := cartDB.Begin()
        defer tx.Rollback()
        var cur int
        _ = tx.QueryRow("SELECT qty FROM cart_items WHERE cart_id=? AND item_id=?", cartID, p.ID).Scan(&cur)
        if cur > 0 {
            _, err = tx.Exec("UPDATE cart_items SET qty=qty+? WHERE cart_id=? AND item_id=?", p.Qty, cartID, p.ID)
        } else {
            _, err = tx.Exec("INSERT INTO cart_items(cart_id, item_id, title, price, image, qty) VALUES(?,?,?,?,?,?)", cartID, p.ID, p.Title, p.Price, p.Image, p.Qty)
        }
        if err != nil { http.Error(w, err.Error(), 500); return }
        _ = tx.Commit()
        writeJSON(w, map[string]string{"status":"ok"})
    case http.MethodPatch:
        var p struct { ID string `json:"id"`; Qty int `json:"qty"` }
        if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
        if strings.TrimSpace(p.ID) == "" { http.Error(w, "id required", 400); return }
        if p.Qty <= 0 { p.Qty = 1 }
        _, err := cartDB.Exec("UPDATE cart_items SET qty=? WHERE cart_id=? AND item_id=?", p.Qty, cartID, p.ID)
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    case http.MethodDelete:
        id := strings.TrimSpace(r.URL.Query().Get("id"))
        all := strings.TrimSpace(r.URL.Query().Get("all"))
        var err error
        if id == "" && all == "1" {
            _, err = cartDB.Exec("DELETE FROM cart_items WHERE cart_id=?", cartID)
        } else if id != "" {
            _, err = cartDB.Exec("DELETE FROM cart_items WHERE cart_id=? AND item_id=?", cartID, id)
        } else {
            http.Error(w, "id or all=1 required", 400); return
        }
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}


