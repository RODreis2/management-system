package handlers

import (
    "html/template"
    "management-system/db"
    "net/http"
    "time"

    "github.com/google/uuid"
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
    cookie, err := r.Cookie("userUUID")
    loggedIn := err == nil
    isAdmin := false

    if loggedIn {
        userUUID, err := uuid.Parse(cookie.Value)
        if err == nil {
            var adminFlag bool
            err = db.DB.QueryRow("SELECT is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&adminFlag)
            if err == nil {
                isAdmin = adminFlag
            }
        }
    }

    data := PageData{
        Title:    "Welcome",
        Message:  "Hello, user! Please login or register.",
        Year:     time.Now().Year(),
        LoggedIn: loggedIn,
        IsAdmin:  isAdmin,
    }
    Tmpl.ExecuteTemplate(w, "index.html", data)
}