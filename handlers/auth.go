package handlers

import (
    "management-system/db"
    "net/http"
    "time"

    "golang.org/x/crypto/bcrypt"
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
    err := db.DB.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPassword)
    if err != nil || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
        data := PageData{
            Title: "Login",
            Year:  time.Now().Year(),
            Error: "Invalid username or password",
        }
        Tmpl.ExecuteTemplate(w, "login.html", data)
        return
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
            Error: "Error processing registration",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    _, err = db.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashedPassword))
    if err != nil {
        data := PageData{
            Title: "Register",
            Year:  time.Now().Year(),
            Error: "Username already exists or registration failed",
        }
        Tmpl.ExecuteTemplate(w, "register.html", data)
        return
    }

    http.Redirect(w, r, "/login", http.StatusSeeOther)
}