package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "html"
    "net/http"
    "net/url"
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

    // Simple HTML page using existing styles
    fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s</title>
  <link rel="icon" href="/img/icon.ico" type="image/x-icon" />
  <link rel="stylesheet" href="/front/CSS/style.css" />
  <link rel="stylesheet" href="/front/CSS/news.css" />
  <style>.news-article{max-width:900px;margin:40px auto;padding:0 16px}.news-article h1{margin:0 0 12px}.news-article .date{color:#666;margin-bottom:16px}.news-article img{max-width:100%%;border-radius:8px;margin:8px 0 16px}</style>
  <script>window.addEventListener('DOMContentLoaded',function(){var b=document.querySelector('.back-link');if(b){b.addEventListener('click',function(e){e.preventDefault();window.history.length>1?history.back():location.href='/front/HTML/news.html';});}});</script>
  <script>/* cache-bust images */document.addEventListener('DOMContentLoaded',function(){var i=document.querySelector('.news-article img');if(i){var u=i.getAttribute('src');if(u){i.setAttribute('src',u+(u.indexOf('?')>-1?'&':'?')+'v='+(Date.now()));}}});</script>
  <link href="https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500;700&display=swap" rel="stylesheet" />
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" />
  </head>
<body>
  <header class="header"><div class="container header-container"><div class="logo"><a href="/front/HTML/main.html"><img src="/img/logo.png" alt="Металл" /></a></div></div></header>
  <main class="news-article">
    <a href="#" class="back-link">← Назад к новостям</a>
    <h1>%s</h1>
    <div class="date">%s</div>
    %s
    <div class="content">%s</div>
  </main>
  <footer class="footer"><div class="container"><div class="copyright"><p>&copy; 2021 - 2025 Межрегионсталь. Все права защищены.</p></div></div></footer>
</body>
</html>`, title, title, date, imgHTML, n.FullText)
}


