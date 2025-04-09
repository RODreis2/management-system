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
    Users    []struct {
        ID       int
        Username string
        UUID     string
    }
    IsAdmin  bool
}

var Tmpl = template.Must(template.ParseFiles(
    "templates/index.html",
    "templates/login.html",
    "templates/register.html",
    "templates/user.html",
    "templates/admin.html",
))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    _, err := r.Cookie("userUUID")
    loggedIn := err == nil

    cookie, err := r.Cookie("isAdmin")
    isAdmin := err == nil && cookie.Value == "true"

    data := PageData{
        Title:    "Welcome",
        Message:  "Hello, user! Please login or register.",
        Year:     time.Now().Year(),
        LoggedIn: loggedIn,
        IsAdmin:  isAdmin,
    }
    Tmpl.ExecuteTemplate(w, "index.html", data)
}