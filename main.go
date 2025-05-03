package main

import (
    "log"
    "management-system/db"
    "management-system/handlers"
    "net/http"
    "strconv"
    "time"
)

func main() {
    db.InitDB()
    defer db.DB.Close()

    // Serve static files with minimal logging
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fs.ServeHTTP(w, r)
    })))

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
    http.HandleFunc("/create_order", handlers.CreateOrderHandler)
    http.HandleFunc("/orders", handlers.OrdersHandler)
    http.HandleFunc("/closed_orders", handlers.ClosedOrdersHandler)
    http.HandleFunc("/view_order", handlers.ViewOrderHandler)
    http.HandleFunc("/edit_order", handlers.EditOrderHandler)
    http.HandleFunc("/close_order", handlers.CloseOrderHandler)
    http.HandleFunc("/image/", handlers.ServeImageHandler)
    http.HandleFunc("/settings", handlers.SettingsHandler)
    http.HandleFunc("/logo", serveLogoHandler)

    log.Println("Server starting on port 8080...")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("Error starting server:", err)
    }
}

func serveLogoHandler(w http.ResponseWriter, r *http.Request) {
    var logoData []byte
    err := db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err != nil || len(logoData) == 0 {
        http.Redirect(w, r, "/static/logo.png", http.StatusTemporaryRedirect)
        return
    }

    // Determine the image type based on the file header
    contentType := "image/png"
    if len(logoData) > 0 {
        if logoData[0] == 0xFF && logoData[1] == 0xD8 {
            contentType = "image/jpeg"
        } else if logoData[0] == 0x89 && logoData[1] == 0x50 && logoData[2] == 0x4E && logoData[3] == 0x47 {
            contentType = "image/png"
        } else if logoData[0] == 0x47 && logoData[1] == 0x49 && logoData[2] == 0x46 {
            contentType = "image/gif"
        }
    }

    w.Header().Set("Content-Type", contentType)
    w.Header().Set("Content-Length", strconv.Itoa(len(logoData)))
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    w.Header().Set("Pragma", "no-cache")
    w.Header().Set("Expires", "0")
    w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
    _, err = w.Write(logoData)
    if err != nil {
        log.Printf("Error writing logo data to response: %v", err)
    }
}