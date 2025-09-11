package main

import (
    "net/http"
    "net/url"
    "time"
)

const sessionCookieName = "admin_session"
const userSessionCookieName = "user_session"

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

// Public user session helpers
func setUserSession(w http.ResponseWriter, login string) {
    cookie := &http.Cookie{
        Name: userSessionCookieName,
        Value: url.QueryEscape(login),
        Path: "/",
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        Expires: time.Now().Add(7*24*time.Hour),
    }
    http.SetCookie(w, cookie)
}

func clearUserSession(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{Name: userSessionCookieName, Value: "", Path: "/", Expires: time.Unix(0,0)})
}


