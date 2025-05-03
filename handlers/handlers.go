package handlers

import (
    "html/template"
    "management-system/db"
    "net/http"
    "time"
    "database/sql"

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
    Theme    string
    Orders   []struct {
        ID           int
        OrderName    string
        Items        string
        Username     string
        Deadline     string
        NearDeadline bool
    }
    NextToExpire []struct {
        ID           int
        OrderName    string
        Deadline     string
    }
    Order    struct {
        ID        int
        OrderName string
        Items     string
        Username  string
        Closed    bool
        Deadline  string
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
    "templates/settings.html",
))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    loggedIn := err == nil
    isAdmin := false
    theme := "light"
    var nextToExpire []struct {
        ID           int
        OrderName    string
        Deadline     string
    }

    if loggedIn {
        userUUID, err := uuid.Parse(cookie.Value)
        if err == nil {
            var adminFlag bool
            var userTheme string
            err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&adminFlag, &userTheme)
            if err == nil {
                isAdmin = adminFlag
                theme = userTheme

                // Fetch up to 5 orders nearing deadline for "Next to Expire" section, ordered by deadline ascending
                rows, err := db.DB.Query("SELECT id, order_name, deadline FROM orders WHERE closed = FALSE AND deadline IS NOT NULL ORDER BY deadline ASC LIMIT 5")
                if err == nil {
                    defer rows.Close()
                    for rows.Next() {
                        var order struct {
                            ID           int
                            OrderName    string
                            Deadline     string
                        }
                        var deadline sql.NullTime
                        if err := rows.Scan(&order.ID, &order.OrderName, &deadline); err == nil && deadline.Valid {
                            order.Deadline = deadline.Time.Format("2006-01-02 15:04")
                            nextToExpire = append(nextToExpire, order)
                        }
                    }
                }
            }
        }
    }

    data := PageData{
        Title:        "Welcome",
        Message:      "Hello, user! Please login or register.",
        Year:         time.Now().Year(),
        LoggedIn:     loggedIn,
        IsAdmin:      isAdmin,
        Theme:        theme,
        NextToExpire: nextToExpire,
    }
    Tmpl.ExecuteTemplate(w, "index.html", data)
}