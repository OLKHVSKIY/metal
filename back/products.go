package main

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "sort"
    "strconv"
    "strings"
)

// ProductRow is a DB-backed product model
type ProductRow struct {
    ID        int64   `json:"id"`
    Type      string  `json:"type"`
    Name      string  `json:"name"`
    Size      string  `json:"size"`
    Subtype   string  `json:"subtype"`
    Img       string  `json:"img"`
    Price     float64 `json:"price"`
    PricePerTon float64 `json:"price_per_ton"`
    ThicknessMM float64 `json:"thickness_mm"`
    WeightKg    float64 `json:"weight_kg"`
    LengthM     float64 `json:"length_m"`
    InStock   bool    `json:"in_stock"`
    Featured  bool    `json:"featured"`
    SKU       string  `json:"sku"`
    CreatedAt string  `json:"created_at"`
}

var typeSlugOptions = []struct{ Slug, Label string }{
    {"armatura", "Арматура"},
    {"truba-profilnaya", "Труба профильная"},
    {"sortovoy-prokat", "Сортовой прокат"},
    {"truba-kruglaya", "Труба круглая"},
    {"listovoy-prokat", "Листовой прокат"},
    {"profnastil", "Профнастил"},
    {"kovanye-izdeliya", "Кованые изделия"},
    {"shtaketnik-metallicheskiy", "Штакетник металлический"},
    {"setka-metallicheskaia", "Сетка металлическая"},
    {"stroymaterialy", "Стройматериалы"},
    {"zabory", "Заборы"},
    {"krepezh", "Крепеж"},
    {"fitingi", "Фитинги"},
    {"vintovye-svai", "Винтовые сваи"},
    {"zaglushki-dlya-profilnyh-trub", "Заглушки для профильных труб"},
}

func categoryToTypeSlug(category string) string {
    switch category {
    case "rebar":
        return "armatura"
    case "profile-pipe":
        return "truba-profilnaya"
    case "round-pipe":
        return "truba-kruglaya"
    case "sheet":
        return "listovoy-prokat"
    case "profnastil":
        return "profnastil"
    case "beam":
        return "sortovoy-prokat"
    default:
        // Already a type slug?
        return category
    }
}

// Prefer DB-backed products if available; fall back to legacy in-memory slice
func queryProductsFromDB(dbh *sql.DB, productType string) ([]ProductRow, error) {
    if dbh == nil { return nil, sql.ErrConnDone }
    var out []ProductRow
    baseWithAll := "SELECT id, ifnull(type,''), ifnull(name,''), ifnull(size,''), ifnull(subtype,''), ifnull(img,''), ifnull(price,0), ifnull(price_per_ton,0), ifnull(thickness_mm,0), ifnull(weight_kg,0), ifnull(length_m,0), ifnull(in_stock,1), ifnull(created_at,''), ifnull(featured,0), ifnull(sku,'') FROM products"
    baseLegacy := "SELECT id, ifnull(type,''), ifnull(name,''), ifnull(size,''), ifnull(img,''), ifnull(price,0), ifnull(created_at,'') FROM products"
    var rows *sql.Rows
    var err error
    // First try full schema
    if strings.TrimSpace(productType) == "" {
        rows, err = dbh.Query(baseWithAll)
    } else {
        rows, err = dbh.Query(baseWithAll+" WHERE lower(type) = lower(?)", productType)
    }
    if err != nil {
        // Auto-migrate missing columns and retry; fallback to legacy select
        if strings.Contains(strings.ToLower(err.Error()), "no such column: subtype") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN subtype TEXT")
            if strings.TrimSpace(productType) == "" {
                rows, err = dbh.Query(baseWithAll)
            } else {
                rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType)
            }
        }
        if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such column: in_stock") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN in_stock INTEGER NOT NULL DEFAULT 1")
            if strings.TrimSpace(productType) == "" {
                rows, err = dbh.Query(baseWithAll)
            } else {
                rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType)
            }
        }
        if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such column: featured") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN featured INTEGER NOT NULL DEFAULT 0")
            if strings.TrimSpace(productType) == "" {
                rows, err = dbh.Query(baseWithAll)
            } else {
                rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType)
            }
        }
        if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such column: sku") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN sku TEXT")
            if strings.TrimSpace(productType) == "" {
                rows, err = dbh.Query(baseWithAll)
            } else {
                rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType)
            }
        }
        if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such column: price_per_ton") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN price_per_ton REAL")
            if strings.TrimSpace(productType) == "" { rows, err = dbh.Query(baseWithAll) } else { rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType) }
        }
        if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such column: thickness_mm") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN thickness_mm REAL")
            if strings.TrimSpace(productType) == "" { rows, err = dbh.Query(baseWithAll) } else { rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType) }
        }
        if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such column: weight_kg") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN weight_kg REAL")
            if strings.TrimSpace(productType) == "" { rows, err = dbh.Query(baseWithAll) } else { rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType) }
        }
        if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such column: length_m") {
            _, _ = dbh.Exec("ALTER TABLE products ADD COLUMN length_m REAL")
            if strings.TrimSpace(productType) == "" { rows, err = dbh.Query(baseWithAll) } else { rows, err = dbh.Query(baseWithAll+" WHERE type = ?", productType) }
        }
        if err != nil {
            // final fallback to legacy layout without new columns
            if strings.TrimSpace(productType) == "" {
                rows, err = dbh.Query(baseLegacy)
            } else {
                rows, err = dbh.Query(baseLegacy+" WHERE lower(type) = lower(?)", productType)
            }
        }
    }
    if err != nil { return nil, err }
    defer rows.Close()
    // Detect number of columns returned to support both schemas
    cols, _ := rows.Columns()
    if len(cols) >= 15 {
        for rows.Next() {
            var r ProductRow
            var inStockInt int
            var featuredInt int
            if err := rows.Scan(&r.ID, &r.Type, &r.Name, &r.Size, &r.Subtype, &r.Img, &r.Price, &r.PricePerTon, &r.ThicknessMM, &r.WeightKg, &r.LengthM, &inStockInt, &r.CreatedAt, &featuredInt, &r.SKU); err != nil { return nil, err }
            r.InStock = inStockInt == 1
            r.Featured = featuredInt == 1
            out = append(out, r)
        }
    } else if len(cols) >= 14 {
        for rows.Next() {
            var r ProductRow
            var inStockInt int
            var featuredInt int
            if err := rows.Scan(&r.ID, &r.Type, &r.Name, &r.Size, &r.Subtype, &r.Img, &r.Price, &r.PricePerTon, &r.ThicknessMM, &r.WeightKg, &r.LengthM, &inStockInt, &r.CreatedAt, &featuredInt); err != nil { return nil, err }
            r.InStock = inStockInt == 1
            r.Featured = featuredInt == 1
            out = append(out, r)
        }
    } else if len(cols) >= 13 {
        for rows.Next() {
            var r ProductRow
            var inStockInt int
            if err := rows.Scan(&r.ID, &r.Type, &r.Name, &r.Size, &r.Subtype, &r.Img, &r.Price, &r.PricePerTon, &r.ThicknessMM, &r.WeightKg, &r.LengthM, &inStockInt, &r.CreatedAt); err != nil { return nil, err }
            r.InStock = inStockInt == 1
            out = append(out, r)
        }
    } else if len(cols) >= 12 {
        for rows.Next() {
            var r ProductRow
            var inStockInt int
            if err := rows.Scan(&r.ID, &r.Type, &r.Name, &r.Size, &r.Subtype, &r.Img, &r.Price, &r.ThicknessMM, &r.WeightKg, &r.LengthM, &inStockInt, &r.CreatedAt); err != nil { return nil, err }
            r.InStock = inStockInt == 1
            out = append(out, r)
        }
    } else if len(cols) >= 9 {
        for rows.Next() {
            var r ProductRow
            var inStockInt int
            if err := rows.Scan(&r.ID, &r.Type, &r.Name, &r.Size, &r.Subtype, &r.Img, &r.Price, &inStockInt, &r.CreatedAt); err != nil { return nil, err }
            r.InStock = inStockInt == 1
            out = append(out, r)
        }
    } else {
        for rows.Next() {
            var r ProductRow
            if err := rows.Scan(&r.ID, &r.Type, &r.Name, &r.Size, &r.Img, &r.Price, &r.CreatedAt); err != nil { return nil, err }
            r.InStock = true
            out = append(out, r)
        }
    }
    return out, nil
}

// Admin CRUD for products
func adminProductsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        t := strings.TrimSpace(r.URL.Query().Get("type"))
        q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
        sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
        order := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("order")))
        rows, err := queryProductsFromDB(productsDB, t)
        if err != nil { writeJSON(w, []ProductRow{}); return }
        // filter by q
        if q != "" {
            filtered := make([]ProductRow, 0, len(rows))
            for _, it := range rows {
                if strings.Contains(strings.ToLower(it.Name+" "+it.Size), q) { filtered = append(filtered, it) }
            }
            rows = filtered
        }
        // sort
        switch sortBy {
        case "price":
            sort.Slice(rows, func(i, j int) bool { if order=="asc" { return rows[i].Price < rows[j].Price } ; return rows[i].Price > rows[j].Price })
        case "name":
            sort.Slice(rows, func(i, j int) bool { if order=="desc" { return rows[i].Name > rows[j].Name } ; return rows[i].Name < rows[j].Name })
        default:
            // created_at desc by default; preserve DB order which is arbitrary; we sort by ID desc
            sort.Slice(rows, func(i, j int) bool { if order=="asc" { return rows[i].ID < rows[j].ID } ; return rows[i].ID > rows[j].ID })
        }
        writeJSON(w, rows)
    case http.MethodPost:
        var p ProductRow
        if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
        p.Type = strings.TrimSpace(p.Type)
        p.Name = strings.TrimSpace(p.Name)
        p.Size = strings.TrimSpace(p.Size)
        p.Img = strings.TrimSpace(p.Img)
        if p.Type == "" || p.Name == "" { http.Error(w, "type and name required", 400); return }
        inStockInt := 0
        if p.InStock { inStockInt = 1 }
        feat := 0
        if p.Featured { feat = 1 }
        // Generate SKU if absent
        if strings.TrimSpace(p.SKU) == "" {
            base := normalizeTypeSlug(p.Type)
            if base == "" { base = "item" }
            p.SKU = strings.ToUpper(base) + "-" + strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(p.Name)), " ", "-")
        }
        res, err := productsDB.Exec("INSERT INTO products(type, name, size, subtype, img, price, price_per_ton, thickness_mm, weight_kg, length_m, in_stock, featured, sku) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)", p.Type, p.Name, p.Size, strings.TrimSpace(p.Subtype), p.Img, p.Price, p.PricePerTon, p.ThicknessMM, p.WeightKg, p.LengthM, inStockInt, feat, p.SKU)
        if err != nil { http.Error(w, err.Error(), 500); return }
        id, _ := res.LastInsertId()
        p.ID = id
        writeJSON(w, p)
    case http.MethodPatch:
        var p ProductRow
        if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
        if p.ID == 0 { http.Error(w, "id required", 400); return }
        // build update: persist subtype and in_stock always; others when provided
        sets := []string{"subtype=?", "in_stock=?"}
        args := []any{strings.TrimSpace(p.Subtype), func() int { if p.InStock { return 1 }; return 0 }()}
        if strings.TrimSpace(p.Type) != "" { sets = append(sets, "type=?"); args = append(args, strings.TrimSpace(p.Type)) }
        if strings.TrimSpace(p.Name) != "" { sets = append(sets, "name=?"); args = append(args, strings.TrimSpace(p.Name)) }
        if p.Size != "" { sets = append(sets, "size=?"); args = append(args, p.Size) }
        if p.Img != "" { sets = append(sets, "img=?"); args = append(args, p.Img) }
        if p.Price != 0 { sets = append(sets, "price=?"); args = append(args, p.Price) }
        if p.PricePerTon != 0 { sets = append(sets, "price_per_ton=?"); args = append(args, p.PricePerTon) }
        if p.ThicknessMM != 0 { sets = append(sets, "thickness_mm=?"); args = append(args, p.ThicknessMM) }
        if p.WeightKg != 0 { sets = append(sets, "weight_kg=?"); args = append(args, p.WeightKg) }
        if p.LengthM != 0 { sets = append(sets, "length_m=?"); args = append(args, p.LengthM) }
        if strings.TrimSpace(p.SKU) != "" { sets = append(sets, "sku=?"); args = append(args, strings.TrimSpace(p.SKU)) }
        if len(sets) == 0 { writeJSON(w, map[string]string{"status":"noop"}); return }
        args = append(args, p.ID)
        _, err := productsDB.Exec("UPDATE products SET "+strings.Join(sets, ", ")+" WHERE id=?", args...)
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    case http.MethodDelete:
        id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
        if id == 0 { http.Error(w, "id required", 400); return }
        _, err := productsDB.Exec("DELETE FROM products WHERE id=?", id)
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    default:
        http.Error(w, "method not allowed", 405)
    }
}

// Public: featured products
func handleFeaturedProducts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet { http.Error(w, "method not allowed", 405); return }
    rows, err := queryProductsFromDB(productsDB, "")
    if err != nil { writeJSON(w, []any{}); return }
    sort.Slice(rows, func(i, j int) bool { return rows[i].ID > rows[j].ID })
    type apiItem map[string]any
    var out []apiItem
    for _, it := range rows {
        if !it.Featured { continue }
        out = append(out, apiItem{
            "id": it.ID,
            "name": it.Name,
            "img": it.Img,
            "price": it.Price,
            "in_stock": it.InStock,
            "type_slug": normalizeTypeSlug(it.Type),
            "size": it.Size,
        })
    }
    writeJSON(w, out)
}

// Admin: toggle featured flag
func adminFeaturedHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        rows, err := productsDB.Query("SELECT id FROM products WHERE featured=1 ORDER BY id DESC")
        if err != nil { writeJSON(w, []int64{}); return }
        defer rows.Close()
        var ids []int64
        for rows.Next() { var id int64; if err := rows.Scan(&id); err==nil { ids = append(ids, id) } }
        writeJSON(w, ids)
    case http.MethodPost, http.MethodPatch:
        var body struct { ID int64 `json:"id"`; Featured bool `json:"featured"` }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, "bad json", 400); return }
        if body.ID == 0 { http.Error(w, "id required", 400); return }
        _, err := productsDB.Exec("UPDATE products SET featured=? WHERE id=?", func() int { if body.Featured { return 1 } ; return 0 }(), body.ID)
        if err != nil { http.Error(w, err.Error(), 500); return }
        writeJSON(w, map[string]string{"status":"ok"})
    default:
        http.Error(w, "method not allowed", 405)
    }
}

// --- Product type descriptions ---
type typeDescription struct {
    Type        string `json:"type"`
    Description string `json:"description"`
}

func adminProductDescriptionsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        rows, err := productsDB.Query("SELECT type, IFNULL(description,'') FROM product_descriptions")
        if err != nil { writeJSON(w, []typeDescription{}); return }
        defer rows.Close()
        var out []typeDescription
        for rows.Next() {
            var t, d string
            if err := rows.Scan(&t, &d); err == nil { out = append(out, typeDescription{Type: t, Description: d}) }
        }
        writeJSON(w, out)
    case http.MethodPost, http.MethodPatch:
        var p typeDescription
        if err := json.NewDecoder(r.Body).Decode(&p); err != nil { http.Error(w, "bad json", 400); return }
        p.Type = strings.TrimSpace(p.Type)
        if p.Type == "" { http.Error(w, "type required", 400); return }
        // Use INSERT OR REPLACE for broader SQLite compatibility
        _, _ = productsDB.Exec("INSERT OR REPLACE INTO product_descriptions(type, description) VALUES(?, ?)", p.Type, p.Description)
        writeJSON(w, map[string]string{"status":"ok"})
    case http.MethodDelete:
        t := strings.TrimSpace(r.URL.Query().Get("type"))
        if t == "" { http.Error(w, "type required", 400); return }
        _, _ = productsDB.Exec("DELETE FROM product_descriptions WHERE type=?", t)
        writeJSON(w, map[string]string{"status":"ok"})
    default:
        http.Error(w, "method not allowed", 405)
    }
}

func handleGetProductDescription(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet { http.Error(w, "method not allowed", 405); return }
    t := strings.TrimSpace(r.URL.Query().Get("type"))
    var d string
    _ = productsDB.QueryRow("SELECT IFNULL(description,'') FROM product_descriptions WHERE type=?", t).Scan(&d)
    writeJSON(w, map[string]string{"type": t, "description": d})
}

// productMatchesText for DB rows (title-only)
func productTitleMatchesSubcategory(title string, category, sub string) bool {
    if sub == "" { return true }
    t := strings.ToLower(title)
    has := func(words ...string) bool {
        for _, w := range words { if strings.Contains(t, w) { return true } }
        return false
    }
    switch category {
    case "profile-pipe":
        switch sub {
        case "otsinkovannaya": return has("оцинк")
        case "kvadratnaya": return has("квадрат", "кв")
        case "pryamougolnaya": return has("прямоуг")
        }
    case "round-pipe":
        switch sub {
        case "ocinkovannaya": return has("оцинк")
        case "besshovnaya": return has("бесшов")
        case "vgp": return has("водогаз", "вгп")
        case "elektrosvarka": return has("электросвар")
        }
    case "sheet":
        switch sub {
        case "ocinkovanniy": return has("оцинк")
        case "st_goryachekatanyi": return has("горячекатан")
        case "st_holodnokatanyi": return has("холоднокатан")
        case "rifleniy_romb": return has("рифлен", "ромб")
        case "riflenaya_chechevica": return has("чечев")
        case "prosechno-vytyazhnoy": return has("просеч", "вытяж")
        }
    case "profnastil":
        switch sub {
        case "ocinkovannyy": return has("оцинк")
        case "krashennyy": return has("крашен")
        case "dlya-zabora": return has("забор")
        }
    case "rebar":
        switch sub {
        case "a500c": return has("а500", "a500")
        case "a1": return has("a1", "а1", "гладкая")
        case "a400": return has("a400", "а400")
        case "fixatory": return has("фиксатор")
        case "stekloplastikovaya": return has("стеклопласт")
        }
    }
    return true
}

