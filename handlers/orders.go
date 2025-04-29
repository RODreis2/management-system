package handlers

import (
    "management-system/db"
    "net/http"
    "time"
    "github.com/google/uuid"
    "log"
    "io"
    "strconv"
    "github.com/russross/blackfriday/v2"
    "strings"
)

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
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

        var isAdmin bool
        err = db.DB.QueryRow("SELECT is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        data := PageData{
            Title:   "Create Order",
            Year:    time.Now().Year(),
            IsAdmin: isAdmin,
        }
        Tmpl.ExecuteTemplate(w, "create_order.html", data)
        return
    }

    if r.Method == "POST" {
        orderName := r.FormValue("orderName")
        items := r.FormValue("items")

        // Handle multiline input
        items = strings.TrimSpace(items)
        items = strings.ReplaceAll(items, "\r\n", "\n")
        items = strings.ReplaceAll(items, "\r", "\n")

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

        r.ParseMultipartForm(32 << 20) // 32 MB
        files := r.MultipartForm.File["images"]
        for _, fileHeader := range files {
            file, err := fileHeader.Open()
            if err != nil {
                log.Printf("Error opening image file: %v", err)
                data := PageData{
                    Title: "Create Order",
                    Year:  time.Now().Year(),
                    Error: "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "create_order.html", data)
                return
            }
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

    var isAdmin bool
    err = db.DB.QueryRow("SELECT is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    rows, err := db.DB.Query("SELECT o.id, o.order_name, o.items, u.username FROM orders o JOIN users u ON o.user_id = u.id WHERE o.closed = FALSE")
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
        IsAdmin: isAdmin,
    }
    Tmpl.ExecuteTemplate(w, "orders.html", data)
}

func ClosedOrdersHandler(w http.ResponseWriter, r *http.Request) {
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

    var isAdmin bool
    err = db.DB.QueryRow("SELECT is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    rows, err := db.DB.Query("SELECT o.id, o.order_name, o.items, u.username FROM orders o JOIN users u ON o.user_id = u.id WHERE o.closed = TRUE")
    if err != nil {
        http.Error(w, "Error fetching closed orders", http.StatusInternalServerError)
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
        Title:  "Closed Orders",
        Year:   time.Now().Year(),
        Orders: orders,
        IsAdmin: isAdmin,
    }
    Tmpl.ExecuteTemplate(w, "closed_orders.html", data)
}

func ViewOrderHandler(w http.ResponseWriter, r *http.Request) {
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

    var isAdmin bool
    err = db.DB.QueryRow("SELECT is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    orderID := r.URL.Query().Get("id")
    if orderID == "" {
        http.Error(w, "Order ID is required", http.StatusBadRequest)
        return
    }

    var orderIDInt int
    var orderName string
    var items string
    var username string
    var closed bool
    err = db.DB.QueryRow("SELECT o.id, o.order_name, o.items, u.username, o.closed FROM orders o JOIN users u ON o.user_id = u.id WHERE o.id = ?", orderID).Scan(&orderIDInt, &orderName, &items, &username, &closed)
    if err != nil {
        http.Error(w, "Error fetching order", http.StatusInternalServerError)
        return
    }

    // Convert Markdown to HTML
    itemsHTML := string(blackfriday.Run([]byte(items)))

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
        Order: struct {
            ID        int
            OrderName string
            Items     string
            Username  string
            Closed    bool
        }{
            ID:        orderIDInt,
            OrderName: orderName,
            Items:     itemsHTML,
            Username:  username,
            Closed:    closed,
        },
        Images:  images,
        OrderID: orderID,
        IsAdmin: isAdmin,
    }
    Tmpl.ExecuteTemplate(w, "view_order.html", data)
}

func EditOrderHandler(w http.ResponseWriter, r *http.Request) {
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

    var isAdmin bool
    err = db.DB.QueryRow("SELECT is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    orderID := r.URL.Query().Get("id")
    if orderID == "" {
        http.Error(w, "Order ID is required", http.StatusBadRequest)
        return
    }

    var closed bool
    err = db.DB.QueryRow("SELECT closed FROM orders WHERE id = ?", orderID).Scan(&closed)
    if err != nil {
        http.Error(w, "Error fetching order status", http.StatusInternalServerError)
        return
    }
    if closed {
        http.Error(w, "Cannot edit a closed order", http.StatusForbidden)
        return
    }

    if r.Method == "GET" {
        var orderIDInt int
        var orderName string
        var items string
        var username string
        err := db.DB.QueryRow("SELECT o.id, o.order_name, o.items, u.username FROM orders o JOIN users u ON o.user_id = u.id WHERE o.id = ?", orderID).Scan(&orderIDInt, &orderName, &items, &username)
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
            Title:   "Edit Order",
            Year:    time.Now().Year(),
            Order: struct {
                ID        int
                OrderName string
                Items     string
                Username  string
                Closed    bool
            }{
                ID:        orderIDInt,
                OrderName: orderName,
                Items:     items,
                Username:  username,
                Closed:    false, // Since we already checked it's not closed
            },
            Images:  images,
            OrderID: orderID,
            IsAdmin: isAdmin,
        }
        Tmpl.ExecuteTemplate(w, "edit_order.html", data)
        return
    }

    if r.Method == "POST" {
        orderName := r.FormValue("orderName")
        items := r.FormValue("items")

        // Handle multiline input
        items = strings.TrimSpace(items)
        items = strings.ReplaceAll(items, "\r\n", "\n")
        items = strings.ReplaceAll(items, "\r", "\n")

        _, err := db.DB.Exec("UPDATE orders SET order_name = ?, items = ? WHERE id = ?", orderName, items, orderID)
        if err != nil {
            log.Printf("Error updating order in database: %v", err)
            data := PageData{
                Title:   "Edit Order",
                Year:    time.Now().Year(),
                OrderID: orderID,
                Error:   "Error updating order. Please try again.",
                IsAdmin: isAdmin,
            }
            Tmpl.ExecuteTemplate(w, "edit_order.html", data)
            return
        }

        r.ParseMultipartForm(32 << 20) // 32 MB
        files := r.MultipartForm.File["images"]
        for _, fileHeader := range files {
            file, err := fileHeader.Open()
            if err != nil {
                log.Printf("Error opening image file: %v", err)
                data := PageData{
                    Title:   "Edit Order",
                    Year:    time.Now().Year(),
                    OrderID: orderID,
                    Error:   "Error uploading image. Please try again.",
                    IsAdmin: isAdmin,
                }
                Tmpl.ExecuteTemplate(w, "edit_order.html", data)
                return
            }
            defer file.Close()

            imageData, err := io.ReadAll(file)
            if err != nil {
                log.Printf("Error reading image file: %v", err)
                data := PageData{
                    Title:   "Edit Order",
                    Year:    time.Now().Year(),
                    OrderID: orderID,
                    Error:   "Error uploading image. Please try again.",
                    IsAdmin: isAdmin,
                }
                Tmpl.ExecuteTemplate(w, "edit_order.html", data)
                return
            }

            _, err = db.DB.Exec("INSERT INTO order_images (order_id, image_data) VALUES (?, ?)", orderID, imageData)
            if err != nil {
                log.Printf("Error inserting image into database: %v", err)
                data := PageData{
                    Title:   "Edit Order",
                    Year:    time.Now().Year(),
                    OrderID: orderID,
                    Error:   "Error uploading image. Please try again.",
                    IsAdmin: isAdmin,
                }
                Tmpl.ExecuteTemplate(w, "edit_order.html", data)
                return
            }
        }

        http.Redirect(w, r, "/view_order?id="+orderID, http.StatusSeeOther)
    }
}

func CloseOrderHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    orderID := r.FormValue("orderID")
    if orderID == "" {
        http.Error(w, "Order ID is required", http.StatusBadRequest)
        return
    }

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

    _, err = db.DB.Exec("UPDATE orders SET closed = TRUE WHERE id = ?", orderID)
    if err != nil {
        log.Printf("Error closing order: %v", err)
        http.Error(w, "Error closing order", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/orders", http.StatusSeeOther)
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