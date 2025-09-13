package main

import (
    "encoding/json"
    "net/http"
    "strings"
)

type socialLinks struct {
    Telegram string `json:"telegram_link"`
    VK       string `json:"vk_link"`
    WP       string `json:"wp_link"`
}

func readSocial() socialLinks {
    var s socialLinks
    _ = userDB.QueryRow("SELECT COALESCE(telegram_link,''), COALESCE(vk_link,''), COALESCE(wp_link,'') FROM social_links WHERE id=1").Scan(&s.Telegram, &s.VK, &s.WP)
    return s
}

// Public GET: /api/social
func handleSocialPublic(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    writeJSON(w, readSocial())
}

// Admin GET/PATCH: /api/admin/social
func adminSocialHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        writeJSON(w, readSocial())
    case http.MethodPatch, http.MethodPost:
        var p socialLinks
        if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
        p.Telegram = strings.TrimSpace(p.Telegram)
        p.VK = strings.TrimSpace(p.VK)
        p.WP = strings.TrimSpace(p.WP)
        _, _ = userDB.Exec("UPDATE social_links SET telegram_link=?, vk_link=?, wp_link=? WHERE id=1", p.Telegram, p.VK, p.WP)
        writeJSON(w, map[string]string{"status":"ok"})
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}


