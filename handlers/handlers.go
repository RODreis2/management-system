package handlers

import (
    "html/template"
    "net/http"
    "time"
)

type PageData struct {
    Title   string
    Message string
    Year    int
    Error   string
}

var Tmpl = template.Must(template.ParseFiles(
    "templates/index.html",
    "templates/login.html",
    "templates/register.html",
))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    data := PageData{
        Title:   "Welcome",
        Message: "Hello, user! Please login or register.",
        Year:    time.Now().Year(),
    }
    Tmpl.ExecuteTemplate(w, "index.html", data)
}