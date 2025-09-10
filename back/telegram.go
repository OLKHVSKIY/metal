package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
)

// sendTelegram posts a message to the configured Telegram chat (best-effort)
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


