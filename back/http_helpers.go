package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "time"
)

func writeJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    _ = enc.Encode(v)
}

// CSRF helpers
func generateCSRFToken() string {
    ts := strconv.FormatInt(time.Now().Unix(), 10)
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


