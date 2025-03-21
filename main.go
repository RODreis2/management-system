package main

import (
    "html/template"
    "log"
    "net/http"
    "time"
)

// TODO: Alterar esse struct pra gente pegar da database.
type PageData struct {
    Title   string
    Message string
    Year    int
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
    currentYear := time.Now().Year()

    // Also TODO: Alterar isso pra ser um pouco mais personalizavel
    data := PageData{
        Title:   "Inventory System",
        Message: "Welcome to the initial page.",
        Year:    currentYear,
    }

    err := tmpl.Execute(w, data)
    if err != nil {
        log.Println("Error executing template:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    http.HandleFunc("/", indexHandler)

    log.Println("Server starting on port 8080...")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("Error starting server:", err)
    }
}