package handlers

import (
    "management-system/db"
    "net/http"
    "time"
    "log"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var logoURL string
    var logoData []byte
    err := db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    if r.Method == "GET" {
        data := PageData{
            Title:   "Login",
            Year:    time.Now().Year(),
            Theme:   "light",
            LogoURL: logoURL,
        }
        Tmpl.ExecuteTemplate(w, "login.html", data)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")
    log.Printf("Login attempt for username: %s", username)

    var storedPassword string
    var userUUID string
    var isAdmin bool
    err = db.DB.QueryRow("SELECT password, uuid, is_admin FROM users WHERE username = ?", username).Scan(&storedPassword, &userUUID, &isAdmin)
    if err != nil || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
        log.Printf("Login failed for username: %s", username)
        data := PageData{
            Title:   "Login",
            Year:    time.Now().Year(),
            Theme:   "light",
            LogoURL: logoURL,
            Error:   "Invalid username or password",
        }
        Tmpl.ExecuteTemplate(w, "login.html", data)
        return
    }

    // Generate a new UUID for the user
    newUUID := uuid.New().String()
    _, err = db.DB.Exec("UPDATE users SET uuid = ? WHERE username = ?", newUUID, username)
    if err != nil {
        log.Printf("Error updating UUID for username %s: %v", username, err)
        data := PageData{
            Title:   "Login",
            Year:    time.Now().Year(),
            Theme:   "light",
            LogoURL: logoURL,
            Error:   "Error updating user UUID",
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

    log.Printf("Login successful for username: %s, new UUID: %s", username, newUUID)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Check if the user is an admin
        cookie, err := r.Cookie("userUUID")
        if err != nil {
            log.Println("Register access denied: No userUUID cookie found")
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        userUUID, err := uuid.Parse(cookie.Value)
        if err != nil {
            log.Printf("Register access denied: Invalid UUID format in cookie: %v", err)
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        var isAdmin bool
        var theme string
        err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
        if err != nil || !isAdmin {
            log.Printf("Register access denied for UUID %s: Not admin or error fetching data: %v", userUUID.String(), err)
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        // Fetch the logo URL from the database if available
        var logoURL string
        var logoData []byte
        err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
        if err == nil && len(logoData) > 0 {
            logoURL = "/logo"
        }

        data := PageData{
            Title:   "Register",
            Year:    time.Now().Year(),
            Theme:   theme,
            LogoURL: logoURL,
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    // Check if the user is an admin
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("Register POST access denied: No userUUID cookie found")
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("Register POST access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil || !isAdmin {
        log.Printf("Register POST access denied for UUID %s: Not admin or error fetching data: %v", userUUID.String(), err)
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    // Fetch the logo URL from the database if available
    var logoURL string
    var logoData []byte
    err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    username := r.FormValue("username")
    password := r.FormValue("password")
    log.Printf("Admin %s registering new user: %s", userUUID.String(), username)

    // Check if username already exists
    var existingUsername string
    err = db.DB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUsername)
    if err == nil {
        data := PageData{
            Title:   "Register",
            Year:    time.Now().Year(),
            Theme:   theme,
            LogoURL: logoURL,
            Error:   "Username already exists. Please choose a different one.",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("Error hashing password for new user %s: %v", username, err)
        data := PageData{
            Title:   "Register",
            Year:    time.Now().Year(),
            Theme:   theme,
            LogoURL: logoURL,
            Error:   "Error processing registration",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    userUUIDStr := uuid.New().String()
    _, err = db.DB.Exec("INSERT INTO users (username, password, uuid) VALUES (?, ?, ?)", username, string(hashedPassword), userUUIDStr)
    if err != nil {
        log.Printf("Error inserting user %s into database: %v", username, err)
        data := PageData{
            Title:   "Register",
            Year:    time.Now().Year(),
            Theme:   theme,
            LogoURL: logoURL,
            Error:   "Registration failed. Please try again.",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    log.Printf("New user %s registered successfully by admin %s", username, userUUID.String())
    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}