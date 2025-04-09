package main

import (
    "log"
    "management-system/db"
    "management-system/handlers"
    "net/http"
)

func main() {
    db.InitDB()
    defer db.DB.Close()

    http.HandleFunc("/", handlers.IndexHandler)
    http.HandleFunc("/login", handlers.LoginHandler)
    http.HandleFunc("/register", handlers.RegisterHandler)
    http.HandleFunc("/user", handlers.UserDataHandler)
    http.HandleFunc("/logout", handlers.LogoutHandler)
    http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            r.ParseForm()
            if r.FormValue("_method") == "DELETE" {
                r.Method = "DELETE"
            }
        }
        handlers.AdminHandler(w, r)
    })

    log.Println("Server starting on port 8080...")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("Error starting server:", err)
    }
}