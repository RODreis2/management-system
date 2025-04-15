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
    Orders   []struct {
        ID        int
        OrderName string
        Items     string
        Username  string
    }
    Order    struct {
        ID        int
        OrderName string
        Items     string
        Username  string
        Closed    bool
    }
    Images   []string
    OrderID  string
}

var Tmpl = template.Must(template.New("").Funcs(template.FuncMap{
    "safeHTML": func(html string) template.HTML {
        return template.HTML(html)
    },
}).ParseFiles(
    "templates/index.html",
    "templates/login.html",
    "templates/register.html",
    "templates/user.html",
    "templates/admin.html",
    "templates/create_order.html",
    "templates/orders.html",
    "templates/view_order.html",
    "templates/edit_order.html",
    "templates/closed_orders.html",
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