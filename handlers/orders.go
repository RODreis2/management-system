package handlers

import (
    "management-system/db"
    "net/http"
    "time"
    "github.com/google/uuid"
    "log"
    "io"
    "strconv"
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

        result, err := db.DB.Exec("INSERT INTO orders (order_name, items, user_id) VALUES (?, ?, ?)", orderName, items, userID)
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

        orderID, err := result.LastInsertId()
        if err != nil {
            log.Printf("Error getting last insert ID: %v", err)
            data := PageData{
                Title: "Create Order",
                Year:  time.Now().Year(),
                Error: "Error creating order. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "create_order.html", data)
            return
        }

        file, _, err := r.FormFile("image")
        if err == nil {
            defer file.Close()

            imageData, err := io.ReadAll(file)
            if err != nil {
                log.Printf("Error reading image file: %v", err)
                data := PageData{
                    Title: "Create Order",
                    Year:  time.Now().Year(),
                    Error: "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "create_order.html", data)
                return
            }

            _, err = db.DB.Exec("INSERT INTO order_images (order_id, image_data) VALUES (?, ?)", orderID, imageData)
            if err != nil {
                log.Printf("Error inserting image into database: %v", err)
                data := PageData{
                    Title: "Create Order",
                    Year:  time.Now().Year(),
                    Error: "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "create_order.html", data)
                return
            }
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

func ViewOrderHandler(w http.ResponseWriter, r *http.Request) {
    orderID := r.URL.Query().Get("id")
    if orderID == "" {
        http.Error(w, "Order ID is required", http.StatusBadRequest)
        return
    }

    var order struct {
        ID        int
        OrderName string
        Items     string
        Username  string
    }
    err := db.DB.QueryRow("SELECT o.id, o.order_name, o.items, u.username FROM orders o JOIN users u ON o.user_id = u.id WHERE o.id = ?", orderID).Scan(&order.ID, &order.OrderName, &order.Items, &order.Username)
    if err != nil {
        http.Error(w, "Error fetching order", http.StatusInternalServerError)
        return
    }

    rows, err := db.DB.Query("SELECT id FROM order_images WHERE order_id = ?", orderID)
    if err != nil {
        http.Error(w, "Error fetching order images", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var images []string
    for rows.Next() {
        var imageID int
        if err := rows.Scan(&imageID); err != nil {
            http.Error(w, "Error scanning image data", http.StatusInternalServerError)
            return
        }
        images = append(images, "/image/"+strconv.Itoa(imageID))
    }

    data := PageData{
        Title:   "Order Details",
        Year:    time.Now().Year(),
        Order:   order,
        Images:  images,
    }
    Tmpl.ExecuteTemplate(w, "view_order.html", data)
}

func ServeImageHandler(w http.ResponseWriter, r *http.Request) {
    imageID := r.URL.Path[len("/image/"):]
    id, err := strconv.Atoi(imageID)
    if err != nil {
        http.Error(w, "Invalid image ID", http.StatusBadRequest)
        return
    }

    var imageData []byte
    err = db.DB.QueryRow("SELECT image_data FROM order_images WHERE id = ?", id).Scan(&imageData)
    if err != nil {
        http.Error(w, "Error fetching image", http.StatusInternalServerError)
        return
    }

    // Determine the image type based on the file header
    contentType := "image/png"
    if len(imageData) > 0 {
        if imageData[0] == 0xFF && imageData[1] == 0xD8 {
            contentType = "image/jpeg"
        } else if imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E && imageData[3] == 0x47 {
            contentType = "image/png"
        } else if imageData[0] == 0x47 && imageData[1] == 0x49 && imageData[2] == 0x46 {
            contentType = "image/gif"
        }
    }

    w.Header().Set("Content-Type", contentType)
    w.Write(imageData)
}