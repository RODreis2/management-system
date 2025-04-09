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

    if isAdmin {
        http.SetCookie(w, &http.Cookie{
            Name:     "isAdmin",
            Value:    "true",
            Path:     "/",
            HttpOnly: true,
            MaxAge:   3600, // 1 hour
        })
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Check if the user is an admin
        cookie, err := r.Cookie("isAdmin")
        if err != nil || cookie.Value != "true" {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    // Check if the user is an admin
    cookie, err := r.Cookie("isAdmin")
    if err != nil || cookie.Value != "true" {
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
            Error: "Error processing registration",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    userUUID := uuid.New().String()
    _, err = db.DB.Exec("INSERT INTO users (username, password, uuid) VALUES (?, ?, ?)", username, string(hashedPassword), userUUID)
    if err != nil {
        log.Printf("Error inserting user into database: %v", err)
        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
            Error: "Registration failed. Please try again.",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}