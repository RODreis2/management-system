package handlers

import (
    "management-system/db"
    "net/http"
    "time"
    "github.com/google/uuid"
    "log"
)

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        data := PageData{
            Title: "Create Order",
            Year:  time.Now().Year(),
        }
        Tmpl.ExecuteTemplate(w, "create_order.html", data)
        return
    }

    if r.Method == "POST" {
        orderName := r.FormValue("orderName")
        items := r.FormValue("items")

        cookie, err := r.Cookie("userUUID")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        userUUID, err := uuid.Parse(cookie.Value)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        var userID int
        err = db.DB.QueryRow("SELECT id FROM users WHERE uuid = ?", userUUID.String()).Scan(&userID)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        _, err = db.DB.Exec("INSERT INTO orders (order_name, items, user_id) VALUES (?, ?, ?)", orderName, items, userID)
        if err != nil {
            log.Printf("Error inserting order into database: %v", err)
            data := PageData{
                Title: "Create Order",
                Year:  time.Now().Year(),
                Error: "Error creating order. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "create_order.html", data)
            return
        }

        http.Redirect(w, r, "/orders", http.StatusSeeOther)
    }
}

func OrdersHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.DB.Query("SELECT o.id, o.order_name, o.items, u.username FROM orders o JOIN users u ON o.user_id = u.id")
    if err != nil {
        http.Error(w, "Error fetching orders", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var orders []struct {
        ID        int
        OrderName string
        Items     string
        Username  string
    }

    for rows.Next() {
        var order struct {
            ID        int
            OrderName string
            Items     string
            Username  string
        }
        if err := rows.Scan(&order.ID, &order.OrderName, &order.Items, &order.Username); err != nil {
            http.Error(w, "Error scanning order data", http.StatusInternalServerError)
            return
        }
        orders = append(orders, order)
    }

    data := PageData{
        Title:  "All Orders",
        Year:   time.Now().Year(),
        Orders: orders,
    }
    Tmpl.ExecuteTemplate(w, "orders.html", data)
}