package main

import (
    "encoding/json"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "strings"
)

type User struct {
    ID int64 `json:"id"`
    Login string `json:"login"`
    Email string `json:"email"`
    Phone string `json:"phone"`
    IsAdmin bool `json:"is_admin"`
    Password string `json:"password,omitempty"`
}

func adminUsersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        rows, err := userDB.Query("SELECT id, login, email, phone, is_admin FROM users ORDER BY id DESC")
        if err != nil { http.Error(w, err.Error(), 500); return }
        defer rows.Close()
        var out []User
        for rows.Next() {
            var u User
            var adminInt int
            if err := rows.Scan(&u.ID, &u.Login, &u.Email, &u.Phone, &adminInt); err != nil { http.Error(w, err.Error(), 500); return }
            u.IsAdmin = adminInt==1
            out = append(out, u)
        }
        writeJSON(w, out)
    case http.MethodPost:
        var u User
        if err := json.NewDecoder(r.Body).Decode(&u); err != nil { http.Error(w, err.Error(), 400); return }
        if strings.TrimSpace(u.Login)=="" || strings.TrimSpace(u.Password)=="" { http.Error(w, "login and password required", 400); return }
        hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
        adminInt := 0
        if u.IsAdmin { adminInt = 1 }
        res, err := userDB.Exec("INSERT INTO users (login, password_hash, email, phone, is_admin) VALUES (?,?,?,?,?)", u.Login, string(hash), u.Email, u.Phone, adminInt)
        if err != nil { http.Error(w, err.Error(), 500); return }
        id, _ := res.LastInsertId()
        writeJSON(w, map[string]any{"id": id, "status":"ok"})
    case http.MethodPatch:
        var u User
        if err := json.NewDecoder(r.Body).Decode(&u); err != nil { http.Error(w, err.Error(), 400); return }
        if u.ID == 0 { http.Error(w, "id required", 400); return }
        // Build dynamic update
        fields := []string{}
        args := []any{}
        if u.Login != "" { fields = append(fields, "login=?"); args = append(args, u.Login) }
        if u.Email != "" || u.Email == "" { fields = append(fields, "email=?"); args = append(args, u.Email) }
        if u.Phone != "" || u.Phone == "" { fields = append(fields, "phone=?"); args = append(args, u.Phone) }
        if u.Password != "" { hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost); fields = append(fields, "password_hash=?"); args = append(args, string(hash)) }
        adminInt := 0
        if u.IsAdmin { adminInt = 1 }
        fields = append(fields, "is_admin=?")
        args = append(args, adminInt)
        if len(fields)==0 { writeJSON(w, map[string]string{"status":"ok"}); return }
        args = append(args, u.ID)
        q := "UPDATE users SET " + strings.Join(fields, ",") + " WHERE id=?"
        if _, err := userDB.Exec(q, args...); err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    case http.MethodDelete:
        idStr := r.URL.Query().Get("id")
        if idStr == "" { http.Error(w, "id required", 400); return }
        if _, err := userDB.Exec("DELETE FROM users WHERE id=?", idStr); err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    default:
        http.Error(w, "method not allowed", 405)
    }
}


