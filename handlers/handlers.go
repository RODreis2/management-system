package handlers

import (
    "html/template"
    "net/http"
    "time"
)

type PageData struct {
    Title    string
    Message  string
    Year     int
    Error    string
    LoggedIn bool
}

var Tmpl = template.Must(template.ParseFiles(
    "templates/index.html",
    "templates/login.html",
    "templates/register.html",
    "templates/user.html",
))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    _, err := r.Cookie("userUUID")
    loggedIn := err == nil

    data := PageData{
        Title:    "Welcome",
        Message:  "Hello, user! Please login or register.",
        Year:     time.Now().Year(),
        LoggedIn: loggedIn,
    }
    Tmpl.ExecuteTemplate(w, "index.html", data)
}