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
    var storedUUID string
    var isAdmin bool
    err = db.DB.QueryRow("SELECT username, uuid, is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&username, &storedUUID, &isAdmin)
    if err != nil || storedUUID != userUUID.String() {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    data := PageData{
        Title:   "User Data",
        Message: "Welcome, " + username,
        Year:    time.Now().Year(),
        IsAdmin: isAdmin,
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