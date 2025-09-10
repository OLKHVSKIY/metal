package main

import (
    "net/http"
    "net/url"
)

func isAdminRequest(r *http.Request) bool {
    c, err := r.Cookie(sessionCookieName)
    if err != nil || c.Value == "" { return false }
    login, _ := url.QueryUnescape(c.Value)
    var isAdmin int
    if err := userDB.QueryRow("SELECT is_admin FROM users WHERE login=?", login).Scan(&isAdmin); err != nil { return false }
    return isAdmin == 1
}

func requireAdmin(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !isAdminRequest(r) {
            http.Redirect(w, r, "/admin/login", http.StatusFound)
            return
        }
        h(w, r)
    }
}

// CSRF middleware for state-changing admin APIs
func csrfProtect(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
            token := r.Header.Get("X-CSRF-Token")
            if token == "" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
                _ = r.ParseForm()
                token = r.FormValue("csrf")
            }
            if !validateCSRF(token) {
                http.Error(w, "invalid csrf", http.StatusForbidden)
                return
            }
        }
        h(w, r)
    }
}
// no-op


