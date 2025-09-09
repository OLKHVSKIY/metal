package main

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "encoding/hex"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"

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
type Order struct {
    ID        int64     `json:"id"`
    Service   string    `json:"service"`
    Name      string    `json:"name"`
    Phone     string    `json:"phone"`
    Email     string    `json:"email"`
    Status    string    `json:"status"` // active/closed
    CreatedAt time.Time `json:"createdAt"`
}

var db *sql.DB
var userDB *sql.DB

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
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

// --- Auth helpers ---
const sessionCookieName = "admin_session"
func setSession(w http.ResponseWriter, login string, isAdmin bool) {
    if !isAdmin { return }
    cookie := &http.Cookie{
        Name: sessionCookieName,
        Value: url.QueryEscape(login),
        Path: "/",
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        Expires: time.Now().Add(30*time.Minute),
    }
    http.SetCookie(w, cookie)
}
func clearSession(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", Expires: time.Unix(0,0)})
}
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

func writeJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    _ = enc.Encode(v)
}

// --- Security helpers: CSRF, headers, rate limiting ---
var csrfSecret = mustRandomKey()

func mustRandomKey() []byte {
    buf := make([]byte, 32)
    _, err := rand.Read(buf)
    if err != nil { return []byte("fallback-secret-key-please-restart") }
    return buf
}

func generateCSRFToken() string {
    ts := fmt.Sprintf("%d", time.Now().Unix())
    mac := hmac.New(sha256.New, csrfSecret)
    mac.Write([]byte(ts))
    sig := hex.EncodeToString(mac.Sum(nil))
    return ts + ":" + sig
}

func validateCSRF(token string) bool {
    parts := strings.Split(token, ":")
    if len(parts) != 2 { return false }
    tsStr, sig := parts[0], parts[1]
    // token valid for 2 hours
    ts, err := strconv.ParseInt(tsStr, 10, 64)
    if err != nil { return false }
    if time.Since(time.Unix(ts,0)) > 2*time.Hour { return false }
    mac := hmac.New(sha256.New, csrfSecret)
    mac.Write([]byte(tsStr))
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(sig), []byte(expected))
}

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

// Orders Handlers
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
        go sendTelegram(fmt.Sprintf("Новая заявка: %s\nИмя: %s\nТелефон: %s\nEmail: %s", o.Service, o.Name, o.Phone, o.Email))
        writeJSON(w, map[string]any{"id": id, "status": "ok"})
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

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

// Telegram
func sendTelegram(text string) {
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        botToken = defaultTelegramBotToken
    }
    chatID := os.Getenv("TELEGRAM_CHAT_ID")
    if chatID == "" {
        // use default provided chat id first
        chatID = defaultTelegramChatID
        if chatID == "" {
            chatID = getOrDiscoverChatID(botToken)
        }
    }
    if botToken == "" || chatID == "" {
        log.Printf("telegram not configured")
        return
    }
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
    body := map[string]string{
        "chat_id": chatID,
        "text":    text,
        "parse_mode": "HTML",
    }
    buf, _ := json.Marshal(body)
    req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("telegram send error: %v", err)
        return
    }
    _ = resp.Body.Close()
}

// --- Admin Users API ---
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
