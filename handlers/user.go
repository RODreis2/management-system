package handlers

import (
    "management-system/db"
    "net/http"
    "time"
    "log"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
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
    var theme string
    err = db.DB.QueryRow("SELECT username, uuid, is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&username, &storedUUID, &isAdmin, &theme)
    if err != nil || storedUUID != userUUID.String() {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    data := PageData{
        Title:   "User Data",
        Message: "Welcome, " + username,
        Year:    time.Now().Year(),
        IsAdmin: isAdmin,
        Theme:   theme,
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

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
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
    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT username, is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&username, &isAdmin, &theme)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    if r.Method == "POST" {
        // Handle username update
        newUsername := r.FormValue("username")
        if newUsername != "" && newUsername != username {
            var existingUsername string
            err = db.DB.QueryRow("SELECT username FROM users WHERE username = ?", newUsername).Scan(&existingUsername)
            if err == nil {
                data := PageData{
                    Title:   "Settings",
                    Message: "Update your profile settings.",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    Error:   "Username already exists. Please choose a different one.",
                }
                Tmpl.ExecuteTemplate(w, "settings.html", data)
                return
            }
            _, err = db.DB.Exec("UPDATE users SET username = ? WHERE uuid = ?", newUsername, userUUID.String())
            if err != nil {
                log.Printf("Error updating username: %v", err)
                data := PageData{
                    Title:   "Settings",
                    Message: "Update your profile settings.",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    Error:   "Error updating username. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "settings.html", data)
                return
            }
            username = newUsername
        }

        // Handle password update
        currentPassword := r.FormValue("currentPassword")
        newPassword := r.FormValue("newPassword")
        if currentPassword != "" && newPassword != "" {
            var storedPassword string
            err = db.DB.QueryRow("SELECT password FROM users WHERE uuid = ?", userUUID.String()).Scan(&storedPassword)
            if err != nil || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(currentPassword)) != nil {
                data := PageData{
                    Title:   "Settings",
                    Message: "Update your profile settings.",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    Error:   "Current password is incorrect.",
                }
                Tmpl.ExecuteTemplate(w, "settings.html", data)
                return
            }
            hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
            if err != nil {
                log.Printf("Error hashing new password: %v", err)
                data := PageData{
                    Title:   "Settings",
                    Message: "Update your profile settings.",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    Error:   "Error updating password. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "settings.html", data)
                return
            }
            _, err = db.DB.Exec("UPDATE users SET password = ? WHERE uuid = ?", string(hashedPassword), userUUID.String())
            if err != nil {
                log.Printf("Error updating password in database: %v", err)
                data := PageData{
                    Title:   "Settings",
                    Message: "Update your profile settings.",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    Error:   "Error updating password. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "settings.html", data)
                return
            }
        }

        // Handle theme update
        newTheme := r.FormValue("theme")
        if newTheme == "light" || newTheme == "dark" {
            _, err = db.DB.Exec("UPDATE users SET theme = ? WHERE uuid = ?", newTheme, userUUID.String())
            if err != nil {
                log.Printf("Error updating theme: %v", err)
                data := PageData{
                    Title:   "Settings",
                    Message: "Update your profile settings.",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    Error:   "Error updating theme. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "settings.html", data)
                return
            }
            theme = newTheme
        }

        http.Redirect(w, r, "/settings", http.StatusSeeOther)
        return
    }

    data := PageData{
        Title:   "Settings",
        Message: "Update your profile settings.",
        Year:    time.Now().Year(),
        IsAdmin: isAdmin,
        Theme:   theme,
    }
    Tmpl.ExecuteTemplate(w, "settings.html", data)
}