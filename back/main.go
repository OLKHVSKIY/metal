package main

import (
    "crypto/rand"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "sort"
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
var db *sql.DB
var userDB *sql.DB
var newsDB *sql.DB
var articlesDB *sql.DB
var productsDB *sql.DB
var cartDB *sql.DB
var itemOrdersDB *sql.DB

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

    // articles database
    articlesDB, err = sql.Open("sqlite", "file:articles.db?_pragma=journal_mode(WAL)")
    if err != nil { return err }
    _, err = articlesDB.Exec(`
        CREATE TABLE IF NOT EXISTS articles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            short_text TEXT NOT NULL,
            full_text TEXT NOT NULL,
            published_at TEXT NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
    `)
    if err != nil { return err }
    // seed default articles if table is empty
    var ac int
    _ = articlesDB.QueryRow("SELECT COUNT(1) FROM articles").Scan(&ac)
    if ac == 0 {
        if err := seedDefaultArticles(); err != nil { return err }
    }

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
        CREATE TABLE IF NOT EXISTS social_links (
            id INTEGER PRIMARY KEY CHECK (id=1),
            telegram_link TEXT,
            vk_link TEXT,
            wp_link TEXT
        );
        INSERT OR IGNORE INTO social_links(id, telegram_link, vk_link, wp_link) VALUES(1, '', '', '');
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

    // products database
    productsDB, err = sql.Open("sqlite", "file:products.db?_pragma=journal_mode(WAL)")
    if err != nil { return err }
    _, err = productsDB.Exec(`
        CREATE TABLE IF NOT EXISTS products (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            type TEXT NOT NULL,
            name TEXT NOT NULL,
            size TEXT,
            subtype TEXT,
            img TEXT,
            price REAL NOT NULL DEFAULT 0,
            price_per_ton REAL,
            thickness_mm REAL,
            weight_kg REAL,
            length_m REAL,
            in_stock INTEGER NOT NULL DEFAULT 1,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            featured INTEGER NOT NULL DEFAULT 0
        );
        CREATE INDEX IF NOT EXISTS idx_products_type ON products(type);
        CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
        CREATE INDEX IF NOT EXISTS idx_products_created ON products(created_at DESC);
    `)
    if err != nil { return err }
    // Try to add new columns for older DBs; ignore errors if they exist
    _, _ = productsDB.Exec("ALTER TABLE products ADD COLUMN in_stock INTEGER NOT NULL DEFAULT 1")
    _, _ = productsDB.Exec("ALTER TABLE products ADD COLUMN subtype TEXT")
    _, _ = productsDB.Exec("ALTER TABLE products ADD COLUMN price_per_ton REAL")
    _, _ = productsDB.Exec("ALTER TABLE products ADD COLUMN thickness_mm REAL")
    _, _ = productsDB.Exec("ALTER TABLE products ADD COLUMN weight_kg REAL")
    _, _ = productsDB.Exec("ALTER TABLE products ADD COLUMN length_m REAL")
    _, _ = productsDB.Exec("ALTER TABLE products ADD COLUMN featured INTEGER NOT NULL DEFAULT 0")
    // ensure subtype index exists after potential migration
    _, _ = productsDB.Exec("CREATE INDEX IF NOT EXISTS idx_products_subtype ON products(subtype)")
    // ensure featured index exists (after column may be added above)
    _, _ = productsDB.Exec("CREATE INDEX IF NOT EXISTS idx_products_featured ON products(featured)")
    // descriptions table
    _, _ = productsDB.Exec(`CREATE TABLE IF NOT EXISTS product_descriptions (type TEXT PRIMARY KEY, description TEXT)`)

    // Seed initial sample products from in-memory list if table is empty
    var pcnt int
    _ = productsDB.QueryRow("SELECT COUNT(1) FROM products").Scan(&pcnt)
    if pcnt == 0 {
        for _, p := range products {
            tp := categoryToTypeSlug(p.CategoryID)
            name := p.Title
            size := p.Description
            img := p.Image
            _, _ = productsDB.Exec("INSERT INTO products(type, name, size, img, price, in_stock) VALUES(?,?,?,?,?,1)", tp, name, size, img, 0)
        }
        log.Printf("Seeded %d sample products into products.db", len(products))
    }

    // cart database
    cartDB, err = sql.Open("sqlite", "file:cart.db?_pragma=journal_mode(WAL)")
    if err != nil { return err }
    _, err = cartDB.Exec(`
        CREATE TABLE IF NOT EXISTS cart_items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            cart_id TEXT NOT NULL,
            item_id TEXT NOT NULL,
            title TEXT,
            price REAL NOT NULL DEFAULT 0,
            image TEXT,
            qty INTEGER NOT NULL DEFAULT 1,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_cart_items_cart ON cart_items(cart_id);
        CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_items_unique ON cart_items(cart_id, item_id);
    `)
    if err != nil { return err }
    // item orders database
    // item orders stored in orders-item.db to match requirement
    itemOrdersDB, err = sql.Open("sqlite", "file:orders-item.db?_pragma=journal_mode(WAL)")
    if err != nil { return err }
    _, err = itemOrdersDB.Exec(`
        CREATE TABLE IF NOT EXISTS item_orders (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            item_id TEXT,
            title TEXT,
            qty INTEGER NOT NULL DEFAULT 1,
            price REAL NOT NULL DEFAULT 0,
            total REAL NOT NULL DEFAULT 0,
            phone TEXT,
            user_login TEXT,
            status TEXT NOT NULL DEFAULT 'Ожидает подтверждения',
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_item_orders_created ON item_orders(created_at DESC);
    `)
    if err != nil { return err }
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
    // Catalog routes (public)
    registerCatalogRoutes(mux)
    
    // API endpoints
    mux.HandleFunc("/api/catalog/categories", withCORS(handleGetCategories))
    mux.HandleFunc("/api/catalog/products", withCORS(handleGetProducts))
    mux.HandleFunc("/api/search", withCORS(handleSearch))
    mux.HandleFunc("/api/gost", withCORS(handleGostList))
    mux.HandleFunc("/api/news", withCORS(handleNewsList))
    mux.HandleFunc("/api/social", withCORS(handleSocialPublic))
    // Public auth
    mux.HandleFunc("/api/register", withCORS(publicRegister))
    mux.HandleFunc("/api/login", withCORS(publicLogin))
    mux.HandleFunc("/api/me", withCORS(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
        c, err := r.Cookie(userSessionCookieName)
        if err != nil || c.Value == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        loginVal, _ := url.QueryUnescape(c.Value)
        var id int64
        var login, email, phone string
        var isAdmin int
        err = userDB.QueryRow("SELECT id, login, email, phone, is_admin FROM users WHERE login=? OR email=? OR phone=?", loginVal, loginVal, loginVal).Scan(&id, &login, &email, &phone, &isAdmin)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        writeJSON(w, map[string]any{"id": id, "login": login, "email": email, "phone": phone, "is_admin": isAdmin==1})
    }))
    mux.HandleFunc("/api/logout", withCORS(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
        clearUserSession(w)
        writeJSON(w, map[string]string{"status":"ok"})
    }))
    mux.HandleFunc("/api/articles", withCORS(handleArticlesList))
    mux.HandleFunc("/api/catalog/product_description", withCORS(handleGetProductDescription))
    // Featured products
    mux.HandleFunc("/api/featured", withCORS(handleFeaturedProducts))
    mux.HandleFunc("/api/health", withCORS(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        _, _ = w.Write([]byte("{\n  \"status\": \"ok\"\n}"))
    }))

    // Simple image proxy for external news images to avoid hotlinking and expiring URLs
    mux.HandleFunc("/api/image-proxy", withCORS(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        raw := r.URL.Query().Get("u")
        if strings.TrimSpace(raw) == "" {
            http.Error(w, "missing url", http.StatusBadRequest)
            return
        }
        u, err := url.Parse(raw)
        if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
            http.Error(w, "invalid url", http.StatusBadRequest)
            return
        }
        client := &http.Client{ Timeout: 10 * time.Second }
        req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
        // Pretend to be a browser to avoid some CDNs blocking requests
        req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ImageProxy/1.0)")
        resp, err := client.Do(req)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadGateway)
            return
        }
        defer resp.Body.Close()
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            http.Error(w, fmt.Sprintf("upstream status %d", resp.StatusCode), http.StatusBadGateway)
            return
        }
        ct := resp.Header.Get("Content-Type")
        if ct == "" { ct = "image/jpeg" }
        w.Header().Set("Content-Type", ct)
        w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
        w.Header().Set("Pragma", "no-cache")
        w.Header().Set("Expires", "0")
        // Stream body
        _, _ = io.Copy(w, resp.Body)
    }))

    // Orders API
    mux.HandleFunc("/api/orders", withCORS(ordersHandler))
    mux.HandleFunc("/api/admin/orders", withCORS(csrfProtect(requireAdmin(adminListOrders))))
    mux.HandleFunc("/api/admin/orders/status", withCORS(csrfProtect(requireAdmin(adminUpdateOrderStatus))))
    mux.HandleFunc("/api/admin/users", withCORS(csrfProtect(requireAdmin(adminUsersHandler))))
    mux.HandleFunc("/api/admin/news", withCORS(csrfProtect(requireAdmin(adminNewsHandler))))
    mux.HandleFunc("/api/admin/articles", withCORS(csrfProtect(requireAdmin(adminArticlesHandler))))
    mux.HandleFunc("/api/admin/products", withCORS(csrfProtect(requireAdmin(adminProductsHandler))))
    mux.HandleFunc("/api/admin/featured", withCORS(csrfProtect(requireAdmin(adminFeaturedHandler))))
    mux.HandleFunc("/api/admin/social", withCORS(csrfProtect(requireAdmin(adminSocialHandler))))
    mux.HandleFunc("/api/admin/social/", withCORS(csrfProtect(requireAdmin(adminSocialHandler))))
    mux.HandleFunc("/api/admin/item_orders", withCORS(csrfProtect(requireAdmin(adminItemOrdersList))))
    mux.HandleFunc("/api/admin/item_orders/status", withCORS(csrfProtect(requireAdmin(adminItemOrdersStatus))))

    // Cart API
    mux.HandleFunc("/api/cart", withCORS(cartHandler))
    // Quick buy item order (public)
    mux.HandleFunc("/api/item-order", withCORS(handleCreateItemOrder))
    mux.HandleFunc("/api/item-order/batch", withCORS(handleCreateItemOrderBatch))
    mux.HandleFunc("/api/admin/product_descriptions", withCORS(csrfProtect(requireAdmin(adminProductDescriptionsHandler))))

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
    // Serve images with headers disabling browser cache to ensure fresh updates
    mux.Handle("/img/", http.StripPrefix("/img/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
        w.Header().Set("Pragma", "no-cache")
        w.Header().Set("Expires", "0")
        http.FileServer(http.Dir(imgDirPath)).ServeHTTP(w, r)
    })))

    // Public news page by ID
    mux.HandleFunc("/back/news/", handleNewsPublic)
    mux.Handle("/gost/", http.StripPrefix("/gost/", http.FileServer(http.Dir(gostDirPath))))

    // Root redirect to main page
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.Redirect(w, r, "/front/HTML/main.html", http.StatusFound)
            return
        }
        http.NotFound(w, r)
    })

    // Serve cabinet page from front/HTML
    mux.HandleFunc("/cabinet/", func(w http.ResponseWriter, r *http.Request) {
        p := filepath.Join(frontDirPath, "HTML", "cabinet.html")
        b, err := os.ReadFile(p)
        if err != nil { http.Error(w, "not found", http.StatusNotFound); return }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Header().Set("Cache-Control", "no-store")
        _, _ = w.Write(b)
    })

    // Serve cart page from front/HTML
    mux.HandleFunc("/cart/", func(w http.ResponseWriter, r *http.Request) {
        p := filepath.Join(frontDirPath, "HTML", "cart.html")
        b, err := os.ReadFile(p)
        if err != nil { http.Error(w, "not found", http.StatusNotFound); return }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Header().Set("Cache-Control", "no-store")
        _, _ = w.Write(b)
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
    sub := strings.Trim(strings.ToLower(q.Get("sub")), " ")
    page, _ := strconv.Atoi(q.Get("page"))
    limit, _ := strconv.Atoi(q.Get("limit"))
    if page <= 0 {
        page = 1
    }
    if limit <= 0 || limit > 100 {
        limit = 12
    }

    // Prefer DB-backed products if table has rows for this type
    dbType := categoryToTypeSlug(category)
    if productsDB != nil {
        if rows, err := queryProductsFromDB(productsDB, dbType); err == nil {
            // Fallback: если по типу ничего не нашли (расхождение в данных), пробуем без фильтра типа и отфильтруем вручную
            if len(rows) == 0 && dbType != "" {
                if allRows, err2 := queryProductsFromDB(productsDB, ""); err2 == nil {
                    want := strings.ToLower(strings.TrimSpace(dbType))
                    tmp := make([]ProductRow, 0, len(allRows))
                    for _, it := range allRows {
                        // accept exact match or canonicalized match of type slug
                        if strings.ToLower(strings.TrimSpace(it.Type)) == want || normalizeTypeSlug(it.Type) == want {
                            tmp = append(tmp, it)
                        }
                    }
                    rows = tmp
                }
            }
            // apply subcategory predicate on name
            var arr []ProductRow
            if sub != "" {
                // Unified predicate for belonging to sub
                for _, it := range rows { if itemBelongsToSub(category, sub, it) { arr = append(arr, it) } }
            } else { arr = rows }
            // stable sort: subtype asc, then price asc to match UI expectations
            sort.SliceStable(arr, func(i, j int) bool {
                si := normalizeSubtypeLabel(arr[i].Subtype)
                sj := normalizeSubtypeLabel(arr[j].Subtype)
                if si == sj {
                    return arr[i].Price < arr[j].Price
                }
                return si < sj
            })

            // map to API objects that include price and stock for client rendering
            type apiItem map[string]any
            var out []apiItem
            for _, it := range arr {
                out = append(out, apiItem{
                    "id": strconv.FormatInt(it.ID, 10),
                    "title": it.Name,
                    "image": it.Img,
                    "categoryId": category,
                    "description": it.Size,
                    "name": it.Name,
                    "img": it.Img,
                    "price": it.Price,
                    "price_per_ton": it.PricePerTon,
                    "thickness_mm": it.ThicknessMM,
                    "weight_kg": it.WeightKg,
                    "length_m": it.LengthM,
                    "weight_tons": it.WeightKg/1000.0,
                    "in_stock": it.InStock,
                    "subtype": it.Subtype,
                    "size": it.Size,
                })
            }
            // pagination
            start := (page - 1) * limit
            if start > len(out) { start = len(out) }
            end := start + limit
            if end > len(out) { end = len(out) }
            writeJSON(w, out[start:end])
            return
        }
    }

    var filtered []Product
    for _, p := range products {
        if category != "" && p.CategoryID != category { continue }
        if sub != "" && !productMatchesSubcategory(p, category, sub) { continue }
        filtered = append(filtered, p)
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

// productMatchesSubcategory applies keyword-based matching per category/sub slug
func productMatchesSubcategory(p Product, category, sub string) bool {
    if sub == "" { return true }
    title := strings.ToLower(p.Title + " " + p.Description)
    hasAny := func(keywords ...string) bool {
        for _, k := range keywords {
            if strings.Contains(title, strings.ToLower(k)) { return true }
        }
        return false
    }
    switch category {
    case "profile-pipe":
        switch sub {
        case "otsinkovannaya":
            return hasAny("оцинк")
        case "kvadratnaya":
            return hasAny("квадрат", "кв")
        case "pryamougolnaya":
            return hasAny("прямоуг")
        }
    case "round-pipe":
        switch sub {
        case "ocinkovannaya":
            return hasAny("оцинк")
        case "besshovnaya":
            return hasAny("бесшов")
        case "vgp":
            return hasAny("водогаз", "вгп")
        case "elektrosvarka":
            return hasAny("электросвар")
        }
    case "sheet":
        switch sub {
        case "ocinkovanniy":
            return hasAny("оцинк")
        case "st_goryachekatanyi":
            return hasAny("горячекатан")
        case "st_holodnokatanyi":
            return hasAny("холоднокатан")
        case "rifleniy_romb":
            return hasAny("рифлен", "ромб")
        case "riflenaya_chechevica":
            return hasAny("рифлен", "чечев")
        case "prosechno-vytyazhnoy":
            return hasAny("просечно", "вытяж")
        }
    case "profnastil":
        switch sub {
        case "ocinkovannyy":
            return hasAny("оцинк")
        case "krashennyy":
            return hasAny("крашен")
        case "dlya-zabora":
            return hasAny("забор")
        }
    case "rebar":
        switch sub {
        case "a500c":
            return hasAny("а500", "a500")
        case "a1":
            return hasAny("a1", "а1", "гладкая")
        case "a400":
            return hasAny("a400", "а400")
        case "fixatory":
            return hasAny("фиксатор", "фиксаторы", "фиксаторы для арматуры")
        case "stekloplastikovaya":
            return hasAny("стеклопластик", "стеклопластиковая")
        }
    }
    // default: no strict match -> include
    return true
}

// Map (category, subSlug) -> human label used in DB subtype
func subSlugToLabel(category, sub string) string {
    switch category {
    case "profile-pipe":
        switch sub { case "kvadratnaya": return "Труба квадратная"; case "otsinkovannaya": return "Труба оцинкованная"; case "pryamougolnaya": return "Труба прямоугольная" }
    case "rebar":
        switch sub { case "a500c": return "Арматура А500C"; case "a1": return "Гладкая арматура A1"; case "a400": return "Арматура A400"; case "fixatory": return "Фиксаторы для арматуры"; case "stekloplastikovaya": return "Арматура стеклопластиковая" }
    case "round-pipe":
        switch sub { case "besshovnaya": return "Бесшовные трубы"; case "vgp": return "Труба водогазопроводная"; case "ocinkovannaya": return "Труба оцинкованная"; case "elektrosvarka": return "Труба электросварная" }
    case "sheet":
        switch sub { case "ocinkovanniy": return "Лист оцинкованный"; case "st_goryachekatanyi": return "Лист стальной горячекатаный"; case "st_holodnokatanyi": return "Лист стальной холоднокатаный"; case "rifleniy_romb": return "Лист рифленый ромб"; case "riflenaya_chechevica": return "Лист рифленый чечевица"; case "prosechno-vytyazhnoy": return "Лист просечно-вытяжной" }
    case "profnastil":
        switch sub { case "krashennyy": return "Профнастил крашеный"; case "ocinkovannyy": return "Профнастил оцинкованный"; case "dlya-zabora": return "Профнастил для забора" }
    }
    return ""
}

// normalizeSubtypeLabel normalizes minor Cyrillic/Latin variations (A/А, C/С), spaces and case
func normalizeSubtypeLabel(s string) string {
    // Replace Cyrillic letters with Latin counterparts for A and C which appear in designations
    replacer := strings.NewReplacer(
        "А", "A", "а", "a",
        "С", "C", "с", "c",
    )
    ns := replacer.Replace(s)
    ns = strings.ToLower(strings.TrimSpace(ns))
    // collapse multiple spaces
    for strings.Contains(ns, "  ") { ns = strings.ReplaceAll(ns, "  ", " ") }
    return ns
}

// normalizeTypeSlug extracts canonical slug from products.type
func normalizeTypeSlug(s string) string {
    ns := strings.ToLower(strings.TrimSpace(s))
    // Replace long dash and em dash variants with space, then take first token
    ns = strings.NewReplacer("—", " ", "-", "-", "\u2014", " ").Replace(ns)
    if i := strings.IndexAny(ns, " \t"); i >= 0 { ns = ns[:i] }
    return ns
}

// normalizeCode reduces a string to ascii a-z0-9, mapping common Cyrillic letters to Latin (а->a, с->c)
func normalizeCode(s string) string {
    s = strings.ToLower(strings.TrimSpace(s))
    s = strings.NewReplacer("А", "A", "а", "a", "С", "C", "с", "c").Replace(s)
    b := make([]rune, 0, len(s))
    for _, r := range s {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
            b = append(b, r)
        }
    }
    return string(b)
}

func wantedCodeForSub(category, sub string) string {
    switch category {
    case "rebar":
        switch sub {
        case "a500c":
            return "a500c"
        case "a1":
            return "a1"
        case "a400":
            return "a400"
        default:
            return ""
        }
    default:
        return ""
    }
}

// optional additional keyword hints for other categories/subs
func wantedKeywordsForSub(category, sub string) []string {
    switch category {
    case "profile-pipe":
        switch sub {
        case "kvadratnaya":
            return []string{"kvadrat"}
        case "otsinkovannaya":
            return []string{"ocink"}
        case "pryamougolnaya":
            return []string{"pryamougol"}
        }
    case "round-pipe":
        switch sub {
        case "besshovnaya":
            return []string{"besshov"}
        case "vgp":
            return []string{"vgp", "vodogaz"}
        case "ocinkovannaya":
            return []string{"ocink"}
        case "elektrosvarka":
            return []string{"elektrosvar"}
        }
    case "sheet":
        switch sub {
        case "ocinkovanniy":
            return []string{"ocink"}
        case "st_goryachekatanyi":
            return []string{"goryachekat"}
        case "st_holodnokatanyi":
            return []string{"holodnokat"}
        case "rifleniy_romb":
            return []string{"rifl", "romb"}
        case "riflenaya_chechevica":
            return []string{"chechev"}
        case "prosechno-vytyazhnoy":
            return []string{"prosech", "vytyazh"}
        }
    case "profnastil":
        switch sub {
        case "krashennyy":
            return []string{"krashen", "polimer"}
        case "ocinkovannyy":
            return []string{"ocink"}
        case "dlya-zabora":
            return []string{"zabor"}
        }
    }
    return nil
}

func itemBelongsToSub(category, sub string, it ProductRow) bool {
    if strings.TrimSpace(sub) == "" { return true }
    // exact subtype label match after normalization
    wl := normalizeSubtypeLabel(subSlugToLabel(category, sub))
    if wl != "" && normalizeSubtypeLabel(it.Subtype) == wl { return true }
    // otherwise, keyword/code match on combined text
    wantCode := wantedCodeForSub(category, sub)
    code := normalizeCode(it.Subtype + " " + it.Name + " " + it.Size)
    if wantCode != "" && strings.Contains(code, wantCode) { return true }
    if kws := wantedKeywordsForSub(category, sub); len(kws) > 0 {
        for _, kw := range kws { if strings.Contains(code, kw) { return true } }
    }
    return false
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
