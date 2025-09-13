package main

import (
    "database/sql"
    "encoding/json"
    "html"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
)

// News represents a news article item
type News struct {
    ID          int64  `json:"id"`
    Title       string `json:"title"`
    ShortText   string `json:"short_text"`
    FullText    string `json:"full_text"`
    PublishedAt string `json:"published_at"` // YYYY-MM-DD
    ImageURL    string `json:"image_url"`
}

// Public list of news with optional year filter
func handleNewsList(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    year := strings.TrimSpace(r.URL.Query().Get("year"))
    var rows *sql.Rows
    var err error
    if year == "" {
        rows, err = newsDB.Query("SELECT id, title, short_text, full_text, published_at, COALESCE(image_url,'') FROM news ORDER BY published_at DESC, id DESC")
    } else {
        rows, err = newsDB.Query("SELECT id, title, short_text, full_text, published_at, COALESCE(image_url,'') FROM news WHERE strftime('%Y', published_at)=? ORDER BY published_at DESC, id DESC", year)
    }
    if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
    defer rows.Close()
    var out []News
    for rows.Next() {
        var n News
        if err := rows.Scan(&n.ID, &n.Title, &n.ShortText, &n.FullText, &n.PublishedAt, &n.ImageURL); err != nil { http.Error(w, err.Error(), 500); return }
        out = append(out, n)
    }
    writeJSON(w, out)
}

// Admin CRUD for news
func adminNewsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        rows, err := newsDB.Query("SELECT id, title, short_text, full_text, published_at, COALESCE(image_url,'') FROM news ORDER BY published_at DESC, id DESC")
        if err != nil { http.Error(w, err.Error(), 500); return }
        defer rows.Close()
        var out []News
        for rows.Next() {
            var n News
            if err := rows.Scan(&n.ID, &n.Title, &n.ShortText, &n.FullText, &n.PublishedAt, &n.ImageURL); err != nil { http.Error(w, err.Error(), 500); return }
            out = append(out, n)
        }
        writeJSON(w, out)
    case http.MethodPost:
        var n News
        if err := json.NewDecoder(r.Body).Decode(&n); err != nil { http.Error(w, err.Error(), 400); return }
        n.Title = strings.TrimSpace(n.Title)
        n.ShortText = strings.TrimSpace(n.ShortText)
        n.FullText = strings.TrimSpace(n.FullText)
        if n.Title == "" || n.ShortText == "" || n.FullText == "" { http.Error(w, "title, short_text, full_text required", 400); return }
        if strings.TrimSpace(n.PublishedAt) == "" {
            n.PublishedAt = time.Now().Format("2006-01-02")
        }
        res, err := newsDB.Exec("INSERT INTO news (title, short_text, full_text, published_at, image_url) VALUES (?,?,?,?,?)", n.Title, n.ShortText, n.FullText, n.PublishedAt, strings.TrimSpace(n.ImageURL))
        if err != nil { http.Error(w, err.Error(), 500); return }
        id, _ := res.LastInsertId()
        writeJSON(w, map[string]any{"id": id, "status": "ok"})
    case http.MethodPatch:
        var n News
        if err := json.NewDecoder(r.Body).Decode(&n); err != nil { http.Error(w, err.Error(), 400); return }
        if n.ID == 0 { http.Error(w, "id required", 400); return }
        fields := []string{}
        args := []any{}
        if strings.TrimSpace(n.Title) != "" { fields = append(fields, "title=?"); args = append(args, strings.TrimSpace(n.Title)) }
        // allow clearing/overwriting text fields
        if n.ShortText != "" || n.ShortText == "" { fields = append(fields, "short_text=?"); args = append(args, n.ShortText) }
        if n.FullText != "" || n.FullText == "" { fields = append(fields, "full_text=?"); args = append(args, n.FullText) }
        if strings.TrimSpace(n.PublishedAt) != "" { fields = append(fields, "published_at=?"); args = append(args, strings.TrimSpace(n.PublishedAt)) }
        if n.ImageURL != "" || n.ImageURL == "" { fields = append(fields, "image_url=?"); args = append(args, strings.TrimSpace(n.ImageURL)) }
        if len(fields) == 0 { writeJSON(w, map[string]string{"status":"ok"}); return }
        args = append(args, n.ID)
        q := "UPDATE news SET " + strings.Join(fields, ",") + " WHERE id=?"
        if _, err := newsDB.Exec(q, args...); err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    case http.MethodDelete:
        idStr := r.URL.Query().Get("id")
        if idStr == "" { http.Error(w, "id required", 400); return }
        if _, err := newsDB.Exec("DELETE FROM news WHERE id=?", idStr); err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    default:
        http.Error(w, "method not allowed", 405)
    }
}


// Public single news page: /back/news/{id}
func handleNewsPublic(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    path := strings.TrimPrefix(r.URL.Path, "/back/news/")
    id, err := strconv.ParseInt(strings.TrimSpace(path), 10, 64)
    if err != nil || id <= 0 {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    var n News
    err = newsDB.QueryRow("SELECT id, title, short_text, full_text, published_at, COALESCE(image_url,'') FROM news WHERE id=?", id).
        Scan(&n.ID, &n.Title, &n.ShortText, &n.FullText, &n.PublishedAt, &n.ImageURL)
    if err == sql.ErrNoRows {
        http.NotFound(w, r)
        return
    }
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Build image URL (use proxy for external images)
    img := strings.TrimSpace(n.ImageURL)
    if strings.HasPrefix(strings.ToLower(img), "http://") || strings.HasPrefix(strings.ToLower(img), "https://") {
        img = "/api/image-proxy?u=" + url.QueryEscape(img)
    }
    imgHTML := ""
    if img != "" {
        imgHTML = "<img src=\"" + img + "\" alt=\"\" />"
    }
    // no-cache headers
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
    w.Header().Set("Pragma", "no-cache")
    w.Header().Set("Expires", "0")

    title := html.EscapeString(n.Title)
    date := html.EscapeString(n.PublishedAt)

    // Serve from external template
    tpl, err := os.ReadFile(filepath.Join(frontDirPath, "HTML", "news_item.html"))
    if err != nil { http.Error(w, "template not found", 500); return }
    htmlStr := string(tpl)
    htmlStr = strings.ReplaceAll(htmlStr, "{{TITLE}}", title)
    htmlStr = strings.ReplaceAll(htmlStr, "{{DATE}}", date)
    htmlStr = strings.ReplaceAll(htmlStr, "{{IMG_HTML}}", imgHTML)
    htmlStr = strings.ReplaceAll(htmlStr, "{{CONTENT_HTML}}", n.FullText)
    _, _ = w.Write([]byte(htmlStr))
}


