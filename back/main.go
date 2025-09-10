package main

import (
    "crypto/rand"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    _ "modernc.org/sqlite"
    "golang.org/x/crypto/bcrypt"
)

type Category struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type Product struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Image       string `json:"image"`
    CategoryID  string `json:"categoryId"`
    Description string `json:"description"`
}

var categories = []Category{
    {ID: "rebar", Name: "Арматура"},
    {ID: "profile-pipe", Name: "Труба профильная"},
    {ID: "sheet", Name: "Листовой прокат"},
    {ID: "angle", Name: "Уголок"},
    {ID: "channel", Name: "Швеллер"},
    {ID: "beam", Name: "Балка двутавровая"},
    {ID: "round-pipe", Name: "Труба круглая"},
    {ID: "profnastil", Name: "Профнастил"},
}

var products = []Product{
    {ID: "arm-a500c", Title: "Арматура А500С", Image: "/img/iron1.jpg", CategoryID: "rebar"},
    {ID: "pipe-40x20", Title: "Труба профильная 40x20", Image: "/img/iron2.jpg", CategoryID: "profile-pipe"},
    {ID: "sheet-3mm", Title: "Лист стальной 3мм", Image: "/img/iron3.webp", CategoryID: "sheet"},
    {ID: "angle-50x50", Title: "Уголок стальной 50x50", Image: "/img/iron4.jpg", CategoryID: "angle"},
    {ID: "profnastil-c8", Title: "Профнастил С8", Image: "/img/iron5.webp", CategoryID: "profnastil"},
    {ID: "channel-12p", Title: "Швеллер 12П", Image: "/img/iron6.webp", CategoryID: "channel"},
    {ID: "beam-i", Title: "Балка двутавровая", Image: "/img/iron7.jpg", CategoryID: "beam"},
    {ID: "rebar-a3", Title: "Арматура А3", Image: "/img/iron8.jpg", CategoryID: "rebar"},
    {ID: "round-50mm", Title: "Труба круглая 50мм", Image: "/img/iron9.jpg", CategoryID: "round-pipe"},
    {ID: "sheet-zn", Title: "Лист оцинкованный", Image: "/img/iron10.jpg", CategoryID: "sheet"},
    {ID: "angle-25x25", Title: "Уголок 25x25", Image: "/img/iron11.jpg", CategoryID: "angle"},
    {ID: "profnastil-c20", Title: "Профнастил С20", Image: "/img/iron12.jpg", CategoryID: "profnastil"},
    {ID: "channel-10p", Title: "Швеллер 10П", Image: "/img/iron13.jpg", CategoryID: "channel"},
    {ID: "beam-20b1", Title: "Балка 20Б1", Image: "/img/iron14.jpg", CategoryID: "beam"},
    {ID: "rebar-a800", Title: "Арматура А800", Image: "/img/iron4.jpg", CategoryID: "rebar"},
    {ID: "pipe-60x40", Title: "Труба профильная 60x40", Image: "/img/iron6.webp", CategoryID: "profile-pipe"},
}

var (
    frontDirPath = "front"
    imgDirPath   = "img"
    gostDirPath  = "gost"
)

// Default Telegram configuration
const defaultTelegramBotToken = "7724103152:AAG09wYrm5VpYxlQdN8hqT6npgGVDE3-z_E"
const defaultTelegramChatID = "7257756560"
const telegramChatIDCacheFile = "telegram_chat_id.txt"

// DB and models
var db *sql.DB
var userDB *sql.DB
var newsDB *sql.DB

func initDB() error {
    var err error
    db, err = sql.Open("sqlite", "file:orders.db?_pragma=journal_mode(WAL)")
    if err != nil {
        return err
    }
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS orders (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            service TEXT NOT NULL,
            name TEXT NOT NULL,
            phone TEXT NOT NULL,
            email TEXT,
            status TEXT NOT NULL DEFAULT 'active',
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    `)
    if err != nil { return err }

    // news database
    newsDB, err = sql.Open("sqlite", "file:news.db?_pragma=journal_mode(WAL)")
    if err != nil { return err }
    _, err = newsDB.Exec(`
        CREATE TABLE IF NOT EXISTS news (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            short_text TEXT NOT NULL,
            full_text TEXT NOT NULL,
            published_at TEXT NOT NULL DEFAULT (date('now'))
        );
        CREATE INDEX IF NOT EXISTS idx_news_published_at ON news(published_at DESC);
    `)
    if err != nil { return err }
    // try to add image_url column for older DBs (ignore error if exists)
    _, _ = newsDB.Exec("ALTER TABLE news ADD COLUMN image_url TEXT")

    // users database
    userDB, err = sql.Open("sqlite", "file:users.db?_pragma=journal_mode(WAL)")
    if err != nil { return err }
    _, err = userDB.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            login TEXT NOT NULL UNIQUE,
            password_hash TEXT NOT NULL,
            email TEXT,
            phone TEXT,
            is_admin INTEGER NOT NULL DEFAULT 0
        );
    `)
    if err != nil { return err }
    // seed admin if not exists
    var c int
    _ = userDB.QueryRow("SELECT COUNT(1) FROM users WHERE login='admin'").Scan(&c)
    if c == 0 {
        hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
        _, _ = userDB.Exec("INSERT INTO users (login, password_hash, email, phone, is_admin) VALUES (?,?,?,?,1)",
            "admin", string(hash), "admin@example.com", "+70000000000")
        log.Printf("Seeded default admin user: admin/admin")
    }
    return nil
}

// NOTE: handlers for orders, news, users are defined in orders.go, news.go, users.go
// This file only wires routes and shared helpers.

func dirExists(p string) bool {
    info, err := os.Stat(p)
    if err != nil {
        return false
    }
    return info.IsDir()
}

func resolveDir(name string) string {
    // prefer current working dir
    if dirExists(name) {
        return name
    }
    // try parent dir (run from back/)
    parent := filepath.Join("..", name)
    if dirExists(parent) {
        return parent
    }
    return name
}

func main() {
    if err := initDB(); err != nil {
        log.Fatalf("DB init error: %v", err)
    }
    mux := http.NewServeMux()

    // API endpoints
    mux.HandleFunc("/api/catalog/categories", withCORS(handleGetCategories))
    mux.HandleFunc("/api/catalog/products", withCORS(handleGetProducts))
    mux.HandleFunc("/api/search", withCORS(handleSearch))
    mux.HandleFunc("/api/gost", withCORS(handleGostList))
    mux.HandleFunc("/api/news", withCORS(handleNewsList))
    mux.HandleFunc("/api/health", withCORS(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        _, _ = w.Write([]byte("{\n  \"status\": \"ok\"\n}"))
    }))

    // Orders API
    mux.HandleFunc("/api/orders", withCORS(ordersHandler))
    mux.HandleFunc("/api/admin/orders", withCORS(csrfProtect(requireAdmin(adminListOrders))))
    mux.HandleFunc("/api/admin/orders/status", withCORS(csrfProtect(requireAdmin(adminUpdateOrderStatus))))
    mux.HandleFunc("/api/admin/users", withCORS(csrfProtect(requireAdmin(adminUsersHandler))))
    mux.HandleFunc("/api/admin/news", withCORS(csrfProtect(requireAdmin(adminNewsHandler))))

    // Simple admin UI
    // auth pages
    mux.HandleFunc("/admin/login", loginPage)
    mux.HandleFunc("/admin/logout", logout)
    // protected admin
    mux.HandleFunc("/admin/", requireAdmin(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/admin/" { http.NotFound(w, r); return }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        sendHTMLWithCSRF(w, filepath.Join(frontDirPath, "HTML", "admin.html"))
    }))

    // Resolve static dirs for both running from project root and from back/
    frontDirPath = resolveDir("front")
    imgDirPath = resolveDir("img")
    gostDirPath = resolveDir("gost")
    log.Printf("Static dirs -> front: %s, img: %s, gost: %s", frontDirPath, imgDirPath, gostDirPath)

    // Guard direct access to admin HTML under /front path
    mux.HandleFunc("/front/HTML/admin.html", func(w http.ResponseWriter, r *http.Request) {
        if !isAdminRequest(r) {
            http.Redirect(w, r, "/admin/login", http.StatusFound)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        sendHTMLWithCSRF(w, filepath.Join(frontDirPath, "HTML", "admin.html"))
    })
    mux.HandleFunc("/front/HTML/admin_login.html", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/admin/login", http.StatusFound)
    })

    // Static serving for front/, img/, gost/
    mux.Handle("/front/", http.StripPrefix("/front/", http.FileServer(http.Dir(frontDirPath))))
    mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(imgDirPath))))
    mux.Handle("/gost/", http.StripPrefix("/gost/", http.FileServer(http.Dir(gostDirPath))))

    // Root redirect to main page
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.Redirect(w, r, "/front/HTML/main.html", http.StatusFound)
            return
        }
        http.NotFound(w, r)
    })

    addr := ":8080"
    log.Printf("Server listening on %s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal(err)
    }
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Referrer-Policy", "no-referrer")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        h(w, r)
    }
}

func loginPage(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        sendHTMLWithCSRF(w, filepath.Join(frontDirPath, "HTML", "admin_login.html"))
    case http.MethodPost:
        if err := r.ParseForm(); err != nil { http.Error(w, "bad form", 400); return }
        if !validateCSRF(r.FormValue("csrf")) { http.Error(w, "invalid csrf", 403); return }
        login := strings.TrimSpace(r.FormValue("login"))
        pass := r.FormValue("password")
        var hash string
        var isAdmin int
        err := userDB.QueryRow("SELECT password_hash, is_admin FROM users WHERE login=?", login).Scan(&hash, &isAdmin)
        if err != nil { http.Error(w, "invalid credentials", 401); return }
        if bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)) != nil { http.Error(w, "invalid credentials", 401); return }
        setSession(w, login, isAdmin==1)
        http.Redirect(w, r, "/admin/", http.StatusFound)
    default:
        http.Error(w, "method not allowed", 405)
    }
}

func logout(w http.ResponseWriter, r *http.Request) {
    clearSession(w)
    http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// removed inline login HTML; using external file

// --- Security helpers: CSRF, headers, rate limiting ---
var csrfSecret = mustRandomKey()

func mustRandomKey() []byte {
    buf := make([]byte, 32)
    _, err := rand.Read(buf)
    if err != nil { return []byte("fallback-secret-key-please-restart") }
    return buf
}

// generateCSRFToken and validateCSRF moved to http_helpers.go

func sendHTMLWithCSRF(w http.ResponseWriter, path string) {
    b, err := os.ReadFile(path)
    if err != nil { http.Error(w, "not found", 404); return }
    token := generateCSRFToken()
    html := strings.ReplaceAll(string(b), "{{CSRF_TOKEN}}", token)
    // Security headers
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("Referrer-Policy", "no-referrer")
    _, _ = w.Write([]byte(html))
}

func handleGetCategories(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, categories)
}

func handleGetProducts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    q := r.URL.Query()
    category := q.Get("category")
    page, _ := strconv.Atoi(q.Get("page"))
    limit, _ := strconv.Atoi(q.Get("limit"))
    if page <= 0 {
        page = 1
    }
    if limit <= 0 || limit > 100 {
        limit = 12
    }

    var filtered []Product
    for _, p := range products {
        if category == "" || p.CategoryID == category {
            filtered = append(filtered, p)
        }
    }

    // pagination
    start := (page - 1) * limit
    if start > len(filtered) {
        start = len(filtered)
    }
    end := start + limit
    if end > len(filtered) {
        end = len(filtered)
    }
    writeJSON(w, filtered[start:end])
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
    if q == "" {
        writeJSON(w, []Product{})
        return
    }
    var found []Product
    for _, p := range products {
        if strings.Contains(strings.ToLower(p.Title), q) || strings.Contains(strings.ToLower(p.Description), q) {
            found = append(found, p)
        }
    }
    writeJSON(w, found)
}

func handleGostList(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    // List PDFs under gost/ directory
    entries, err := os.ReadDir(gostDirPath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    type GostFile struct {
        Name string `json:"name"`
        Path string `json:"path"`
    }
    var files []GostFile
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        name := e.Name()
        if strings.HasSuffix(strings.ToLower(name), ".pdf") {
            files = append(files, GostFile{
                Name: name,
                Path: filepath.ToSlash(filepath.Join("/gost", name)),
            })
        }
    }
    writeJSON(w, files)
}

// moved to news.go: handleNewsList, adminNewsHandler

// moved to orders.go: ordersHandler

// moved to orders.go: adminListOrders

// moved to orders.go: adminUpdateOrderStatus

// moved to users.go: type User, adminUsersHandler

// Attempts to read cached chat ID or discover it from getUpdates, then caches it.
func getOrDiscoverChatID(botToken string) string {
    if b, err := os.ReadFile(telegramChatIDCacheFile); err == nil {
        s := strings.TrimSpace(string(b))
        if s != "" {
            return s
        }
    }
    type updatesResp struct {
        OK     bool `json:"ok"`
        Result []struct {
            Message struct {
                Chat struct {
                    ID int64 `json:"id"`
                } `json:"chat"`
            } `json:"message"`
        } `json:"result"`
    }
    u := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", botToken)
    res, err := http.Get(u)
    if err != nil {
        log.Printf("telegram getUpdates error: %v", err)
        return ""
    }
    defer res.Body.Close()
    var ur updatesResp
    if err := json.NewDecoder(res.Body).Decode(&ur); err != nil {
        log.Printf("telegram decode getUpdates error: %v", err)
        return ""
    }
    if !ur.OK || len(ur.Result) == 0 {
        log.Printf("telegram getUpdates has no chats; send a message to the bot to initialize chat")
        return ""
    }
    last := ur.Result[len(ur.Result)-1]
    cid := fmt.Sprintf("%d", last.Message.Chat.ID)
    _ = os.WriteFile(telegramChatIDCacheFile, []byte(cid), 0644)
    return cid
}
