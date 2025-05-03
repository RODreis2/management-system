package handlers

import (
    "management-system/db"
    "net/http"
    "time"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "log"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        data := PageData{
            Title: "Login",
            Year:  time.Now().Year(),
            Theme: "light",
        }
        Tmpl.ExecuteTemplate(w, "login.html", data)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")

    var storedPassword string
    var userUUID string
    var isAdmin bool
    err := db.DB.QueryRow("SELECT password, uuid, is_admin FROM users WHERE username = ?", username).Scan(&storedPassword, &userUUID, &isAdmin)
    if err != nil || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
        data := PageData{
            Title: "Login",
            Year:  time.Now().Year(),
            Theme: "light",
            Error: "Invalid username or password",
        }
        Tmpl.ExecuteTemplate(w, "login.html", data)
        return
    }

    // Generate a new UUID for the user
    newUUID := uuid.New().String()
    _, err = db.DB.Exec("UPDATE users SET uuid = ? WHERE username = ?", newUUID, username)
    if err != nil {
        log.Printf("Error updating UUID: %v", err)
        data := PageData{
            Title: "Login",
            Year:  time.Now().Year(),
            Theme: "light",
            Error: "Error updating user UUID",
        }
        Tmpl.ExecuteTemplate(w, "login.html", data)
        return
    }

    // Set the UUID in a cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "userUUID",
        Value:    newUUID,
        Path:     "/",
        HttpOnly: true,
        MaxAge:   3600, // 1 hour
    })

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Check if the user is an admin
        cookie, err := r.Cookie("userUUID")
        if err != nil {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        userUUID, err := uuid.Parse(cookie.Value)
        if err != nil {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        var isAdmin bool
        var theme string
        err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
        if err != nil || !isAdmin {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
            Theme: theme,
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    // Check if the user is an admin
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil || !isAdmin {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")

    // Check if username already exists
    var existingUsername string
    err = db.DB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUsername)
    if err == nil {
        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
            Theme: theme,
            Error: "Username already exists. Please choose a different one.",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("Error hashing password: %v", err)
        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
            Theme: theme,
            Error: "Error processing registration",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    userUUIDStr := uuid.New().String()
    _, err = db.DB.Exec("INSERT INTO users (username, password, uuid) VALUES (?, ?, ?)", username, string(hashedPassword), userUUIDStr)
    if err != nil {
        log.Printf("Error inserting user into database: %v", err)
        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
            Theme: theme,
            Error: "Registration failed. Please try again.",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}