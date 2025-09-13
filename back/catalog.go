package main

import (
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

// catalogSlugToTitle maps URL slugs to human-readable category titles
var catalogSlugToTitle = map[string]string{
    "armatura":                    "Арматура",
    "truba-profilnaya":            "Труба профильная",
    "sortovoy-prokat":             "Сортовой прокат",
    "truba-kruglaya":              "Труба круглая",
    "listovoy-prokat":             "Листовой прокат",
    "profnastil":                  "Профнастил",
    "kovanye-izdeliya":            "Кованые изделия",
    "shtaketnik-metallicheskiy":   "Штакетник металлический",
    "setka-metallicheskaia":       "Сетка металлическая",
    "stroymaterialy":              "Стройматериалы",
    "zabory":                      "Заборы",
    "krepezh":                     "Крепеж",
    "petli":                       "Петли",
    "fitingi":                     "Фитинги",
    "vintovye-svai":               "Винтовые сваи",
    "zaglushki-dlya-profilnyh-trub": "Заглушки для профильных труб",
}

// registerCatalogRoutes wires handlers for /catalog/ and category pages /catalog/{slug}/
func registerCatalogRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/catalog/", func(w http.ResponseWriter, r *http.Request) {
        // Serve catalog grid on exact /catalog/ (or /catalog)
        if r.URL.Path == "/catalog/" || r.URL.Path == "/catalog" {
            p := filepath.Join(frontDirPath, "HTML", "catalog.html")
            b, err := os.ReadFile(p)
            if err != nil {
                http.Error(w, "not found", http.StatusNotFound)
                return
            }
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.Header().Set("Cache-Control", "no-store")
            _, _ = w.Write(b)
            return
        }

        // Expect /catalog/{slug}/
        trimmed := strings.TrimPrefix(r.URL.Path, "/catalog/")
        trimmed = strings.Trim(trimmed, "/")
        parts := strings.Split(trimmed, "/")
        // if product deep route like /catalog/{type}/{sub}/{name}/{size}/ or /catalog/{type}/{name}/{size}/ -> serve item page html
        if len(parts) >= 3 {
            p := filepath.Join(frontDirPath, "HTML", "catalog_item.html")
            b, err := os.ReadFile(p)
            if err != nil {
                http.Error(w, "not found", http.StatusNotFound)
                return
            }
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.Header().Set("Cache-Control", "no-store")
            _, _ = w.Write(b)
            return
        }

        if len(parts) < 1 || len(parts) > 2 || parts[0] == "" {
            http.NotFound(w, r)
            return
        }
        slug := parts[0]
        sub := ""
        if len(parts) == 2 { sub = parts[1] }
        title, ok := catalogSlugToTitle[slug]
        if !ok {
            http.NotFound(w, r)
            return
        }
        // Serve shared HTML template catalog_list.html and inject title/slug placeholders
        p := filepath.Join(frontDirPath, "HTML", "catalog_list.html")
        b, err := os.ReadFile(p)
        if err != nil {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        html := string(b)
        html = strings.ReplaceAll(html, "{{CATEGORY_TITLE}}", title)
        html = strings.ReplaceAll(html, "{{CATEGORY_SLUG}}", slug)
        html = strings.ReplaceAll(html, "{{SUBCATEGORY_SLUG}}", sub)
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Header().Set("Cache-Control", "no-store")
        _, _ = w.Write([]byte(html))
    })
}


