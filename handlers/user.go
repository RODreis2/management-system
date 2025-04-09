package handlers

import (
    "management-system/db"
    "net/http"
    "time"

    "github.com/google/uuid"
)

func UserDataHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    var username string
    err = db.DB.QueryRow("SELECT username FROM users WHERE uuid = ?", userUUID.String()).Scan(&username)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    data := PageData{
        Title:   "User Data",
        Message: "Welcome, " + username,
        Year:    time.Now().Year(),
    }
    Tmpl.ExecuteTemplate(w, "user.html", data)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    cookie := &http.Cookie{
        Name:   "userUUID",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    }
    http.SetCookie(w, cookie)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}