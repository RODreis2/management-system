package handlers

import (
    "database/sql"
    "management-system/db"
    "net/http"
    "time"
    "log"
    "io"
    "strconv"
    "github.com/google/uuid"
    "github.com/russross/blackfriday/v2"
    "strings"
)

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("Create order access denied: No userUUID cookie found")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("Create order access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil {
        log.Printf("Create order access denied for UUID %s: Error fetching data: %v", userUUID.String(), err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // Fetch the logo URL from the database if available
    var logoURL string
    var logoData []byte
    err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    if r.Method == "GET" {
        data := PageData{
            Title:   "Create Order",
            Year:    time.Now().Year(),
            IsAdmin: isAdmin,
            Theme:   theme,
            LogoURL: logoURL,
        }
        Tmpl.ExecuteTemplate(w, "create_order.html", data)
        return
    }

    if r.Method == "POST" {
        orderName := r.FormValue("orderName")
        items := r.FormValue("items")
        deadlineStr := r.FormValue("deadline")
        log.Printf("User %s creating order: %s", userUUID.String(), orderName)

        // Handle multiline input
        items = strings.TrimSpace(items)
        items = strings.ReplaceAll(items, "\r\n", "\n")
        items = strings.ReplaceAll(items, "\r", "\n")

        var deadline interface{}
        if deadlineStr != "" {
            parsedDeadline, err := time.Parse("2006-01-02T15:04", deadlineStr)
            if err != nil {
                log.Printf("Error parsing deadline for order %s: %v", orderName, err)
                data := PageData{
                    Title:   "Create Order",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Invalid deadline format. Please use YYYY-MM-DD HH:MM.",
                }
                Tmpl.ExecuteTemplate(w, "create_order.html", data)
                return
            }
            deadline = parsedDeadline
        } else {
            deadline = nil
        }

        var userID int
        err = db.DB.QueryRow("SELECT id FROM users WHERE uuid = ?", userUUID.String()).Scan(&userID)
        if err != nil {
            log.Printf("Error fetching user ID for UUID %s: %v", userUUID.String(), err)
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        result, err := db.DB.Exec("INSERT INTO orders (order_name, items, user_id, deadline) VALUES (?, ?, ?, ?)", orderName, items, userID, deadline)
        if err != nil {
            log.Printf("Error inserting order %s into database: %v", orderName, err)
            data := PageData{
                Title:   "Create Order",
                Year:    time.Now().Year(),
                IsAdmin: isAdmin,
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Error creating order. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "create_order.html", data)
            return
        }

        orderID, err := result.LastInsertId()
        if err != nil {
            log.Printf("Error getting last insert ID for order %s: %v", orderName, err)
            data := PageData{
                Title:   "Create Order",
                Year:    time.Now().Year(),
                IsAdmin: isAdmin,
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Error creating order. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "create_order.html", data)
            return
        }

        r.ParseMultipartForm(32 << 20) // 32 MB
        files := r.MultipartForm.File["images"]
        for _, fileHeader := range files {
            file, err := fileHeader.Open()
            if err != nil {
                log.Printf("Error opening image file for order ID %d: %v", orderID, err)
                data := PageData{
                    Title:   "Create Order",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "create_order.html", data)
                return
            }
            defer file.Close()

            imageData, err := io.ReadAll(file)
            if err != nil {
                log.Printf("Error reading image file for order ID %d: %v", orderID, err)
                data := PageData{
                    Title:   "Create Order",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "create_order.html", data)
                return
            }

            _, err = db.DB.Exec("INSERT INTO order_images (order_id, image_data) VALUES (?, ?)", orderID, imageData)
            if err != nil {
                log.Printf("Error inserting image into database for order ID %d: %v", orderID, err)
                data := PageData{
                    Title:   "Create Order",
                    Year:    time.Now().Year(),
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "create_order.html", data)
                return
            }
        }

        log.Printf("Order %s (ID: %d) created successfully by user %s", orderName, orderID, userUUID.String())
        http.Redirect(w, r, "/orders", http.StatusSeeOther)
    }
}

func OrdersHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("Orders access denied: No userUUID cookie found")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("Orders access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil {
        log.Printf("Orders access denied for UUID %s: Error fetching data: %v", userUUID.String(), err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // Fetch the logo URL from the database if available
    var logoURL string
    var logoData []byte
    err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    rows, err := db.DB.Query("SELECT o.id, o.order_name, o.items, u.username, o.deadline FROM orders o JOIN users u ON o.user_id = u.id WHERE o.closed = FALSE")
    if err != nil {
        log.Printf("Error fetching orders for user %s: %v", userUUID.String(), err)
        http.Error(w, "Error fetching orders", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var orders []struct {
        ID           int
        OrderName    string
        Items        string
        Username     string
        Deadline     string
        NearDeadline bool
    }

    currentTime := time.Now()
    for rows.Next() {
        var order struct {
            ID           int
            OrderName    string
            Items        string
            Username     string
            Deadline     string
            NearDeadline bool
        }
        var deadline sql.NullTime
        if err := rows.Scan(&order.ID, &order.OrderName, &order.Items, &order.Username, &deadline); err != nil {
            log.Printf("Error scanning order data for user %s: %v", userUUID.String(), err)
            http.Error(w, "Error scanning order data", http.StatusInternalServerError)
            return
        }
        if deadline.Valid {
            order.Deadline = deadline.Time.Format("2006-01-02 15:04")
            if deadline.Time.Sub(currentTime).Hours()/24 <= 10 {
                order.NearDeadline = true
            }
        } else {
            order.Deadline = "Not Set"
            order.NearDeadline = false
        }
        orders = append(orders, order)
    }

    log.Printf("User %s viewed all orders, count: %d", userUUID.String(), len(orders))
    data := PageData{
        Title:    "All Orders",
        Year:     time.Now().Year(),
        Orders:   orders,
        IsAdmin:  isAdmin,
        Theme:    theme,
        LogoURL:  logoURL,
    }
    Tmpl.ExecuteTemplate(w, "orders.html", data)
}

func ClosedOrdersHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("Closed orders access denied: No userUUID cookie found")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("Closed orders access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil {
        log.Printf("Closed orders access denied for UUID %s: Error fetching data: %v", userUUID.String(), err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // Fetch the logo URL from the database if available
    var logoURL string
    var logoData []byte
    err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    rows, err := db.DB.Query("SELECT o.id, o.order_name, o.items, u.username, o.deadline FROM orders o JOIN users u ON o.user_id = u.id WHERE o.closed = TRUE")
    if err != nil {
        log.Printf("Error fetching closed orders for user %s: %v", userUUID.String(), err)
        http.Error(w, "Error fetching closed orders", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var closedOrders []struct {
        ID           int
        OrderName    string
        Items        string
        Username     string
        Deadline     string
        NearDeadline bool
    }

    for rows.Next() {
        var order struct {
            ID           int
            OrderName    string
            Items        string
            Username     string
            Deadline     string
            NearDeadline bool
        }
        var deadline sql.NullTime
        if err := rows.Scan(&order.ID, &order.OrderName, &order.Items, &order.Username, &deadline); err != nil {
            log.Printf("Error scanning closed order data for user %s: %v", userUUID.String(), err)
            http.Error(w, "Error scanning order data", http.StatusInternalServerError)
            return
        }
        if deadline.Valid {
            order.Deadline = deadline.Time.Format("2006-01-02 15:04")
        } else {
            order.Deadline = "Not Set"
        }
        order.NearDeadline = false // Not needed for closed orders
        closedOrders = append(closedOrders, order)
    }

    log.Printf("User %s viewed closed orders, count: %d", userUUID.String(), len(closedOrders))
    data := PageData{
        Title:    "Closed Orders",
        Year:     time.Now().Year(),
        Orders:   closedOrders,
        IsAdmin:  isAdmin,
        Theme:    theme,
        LogoURL:  logoURL,
    }
    Tmpl.ExecuteTemplate(w, "closed_orders.html", data)
}

func ViewOrderHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("View order access denied: No userUUID cookie found")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("View order access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil {
        log.Printf("View order access denied for UUID %s: Error fetching data: %v", userUUID.String(), err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // Fetch the logo URL from the database if available
    var logoURL string
    var logoData []byte
    err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    orderID := r.URL.Query().Get("id")
    if orderID == "" {
        log.Println("View order failed: No order ID provided")
        http.Error(w, "Order ID is required", http.StatusBadRequest)
        return
    }

    var orderIDInt int
    var orderName string
    var items string
    var username string
    var closed bool
    var deadline sql.NullTime
    err = db.DB.QueryRow("SELECT o.id, o.order_name, o.items, u.username, o.closed, o.deadline FROM orders o JOIN users u ON o.user_id = u.id WHERE o.id = ?", orderID).Scan(&orderIDInt, &orderName, &items, &username, &closed, &deadline)
    if err != nil {
        log.Printf("Error fetching order ID %s for user %s: %v", orderID, userUUID.String(), err)
        http.Error(w, "Error fetching order", http.StatusInternalServerError)
        return
    }

    // Convert Markdown to HTML
    itemsHTML := string(blackfriday.Run([]byte(items)))

    rows, err := db.DB.Query("SELECT id FROM order_images WHERE order_id = ?", orderID)
    if err != nil {
        log.Printf("Error fetching images for order ID %s: %v", orderID, err)
        http.Error(w, "Error fetching order images", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var images []string
    for rows.Next() {
        var imageID int
        if err := rows.Scan(&imageID); err != nil {
            log.Printf("Error scanning image data for order ID %s: %v", orderID, err)
            http.Error(w, "Error scanning image data", http.StatusInternalServerError)
            return
        }
        images = append(images, "/image/"+strconv.Itoa(imageID))
    }

    deadlineStr := "Not Set"
    if deadline.Valid {
        deadlineStr = deadline.Time.Format("2006-01-02 15:04")
    }

    log.Printf("User %s viewed order ID %s: %s", userUUID.String(), orderID, orderName)
    data := PageData{
        Title:   "Order Details",
        Year:    time.Now().Year(),
        Order: struct {
            ID        int
            OrderName string
            Items     string
            Username  string
            Closed    bool
            Deadline  string
        }{
            ID:        orderIDInt,
            OrderName: orderName,
            Items:     itemsHTML,
            Username:  username,
            Closed:    closed,
            Deadline:  deadlineStr,
        },
        Images:  images,
        OrderID: orderID,
        IsAdmin: isAdmin,
        Theme:   theme,
        LogoURL: logoURL,
    }
    Tmpl.ExecuteTemplate(w, "view_order.html", data)
}

func EditOrderHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("Edit order access denied: No userUUID cookie found")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("Edit order access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil {
        log.Printf("Edit order access denied for UUID %s: Error fetching data: %v", userUUID.String(), err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // Fetch the logo URL from the database if available
    var logoURL string
    var logoData []byte
    err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    orderID := r.URL.Query().Get("id")
    if orderID == "" {
        log.Println("Edit order failed: No order ID provided")
        http.Error(w, "Order ID is required", http.StatusBadRequest)
        return
    }

    var closed bool
    err = db.DB.QueryRow("SELECT closed FROM orders WHERE id = ?", orderID).Scan(&closed)
    if err != nil {
        log.Printf("Error fetching order status for ID %s: %v", orderID, err)
        http.Error(w, "Error fetching order status", http.StatusInternalServerError)
        return
    }
    if closed {
        log.Printf("Edit order denied for ID %s: Order is closed", orderID)
        http.Error(w, "Cannot edit a closed order", http.StatusForbidden)
        return
    }

    if r.Method == "GET" {
        var orderIDInt int
        var orderName string
        var items string
        var username string
        var deadline sql.NullTime
        err := db.DB.QueryRow("SELECT o.id, o.order_name, o.items, u.username, o.deadline FROM orders o JOIN users u ON o.user_id = u.id WHERE o.id = ?", orderID).Scan(&orderIDInt, &orderName, &items, &username, &deadline)
        if err != nil {
            log.Printf("Error fetching order ID %s for editing: %v", orderID, err)
            http.Error(w, "Error fetching order", http.StatusInternalServerError)
            return
        }

        deadlineStr := ""
        if deadline.Valid {
            deadlineStr = deadline.Time.Format("2006-01-02T15:04")
        }

        rows, err := db.DB.Query("SELECT id FROM order_images WHERE order_id = ?", orderID)
        if err != nil {
            log.Printf("Error fetching images for order ID %s: %v", orderID, err)
            http.Error(w, "Error fetching order images", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var images []string
        for rows.Next() {
            var imageID int
            if err := rows.Scan(&imageID); err != nil {
                log.Printf("Error scanning image data for order ID %s: %v", orderID, err)
                http.Error(w, "Error scanning image data", http.StatusInternalServerError)
                return
            }
            images = append(images, "/image/"+strconv.Itoa(imageID))
        }

        log.Printf("User %s editing order ID %s: %s", userUUID.String(), orderID, orderName)
        data := PageData{
            Title:   "Edit Order",
            Year:    time.Now().Year(),
            Order: struct {
                ID        int
                OrderName string
                Items     string
                Username  string
                Closed    bool
                Deadline  string
            }{
                ID:        orderIDInt,
                OrderName: orderName,
                Items:     items,
                Username:  username,
                Closed:    false,
                Deadline:  deadlineStr,
            },
            Images:  images,
            OrderID: orderID,
            IsAdmin: isAdmin,
            Theme:   theme,
            LogoURL: logoURL,
        }
        Tmpl.ExecuteTemplate(w, "edit_order.html", data)
        return
    }

    if r.Method == "POST" {
        orderName := r.FormValue("orderName")
        items := r.FormValue("items")
        deadlineStr := r.FormValue("deadline")
        log.Printf("User %s updating order ID %s: %s", userUUID.String(), orderID, orderName)

        // Handle multiline input
        items = strings.TrimSpace(items)
        items = strings.ReplaceAll(items, "\r\n", "\n")
        items = strings.ReplaceAll(items, "\r", "\n")

        var deadline interface{}
        if deadlineStr != "" {
            parsedDeadline, err := time.Parse("2006-01-02T15:04", deadlineStr)
            if err != nil {
                log.Printf("Error parsing deadline for order ID %s: %v", orderID, err)
                data := PageData{
                    Title:   "Edit Order",
                    Year:    time.Now().Year(),
                    OrderID: orderID,
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Invalid deadline format. Please use YYYY-MM-DD HH:MM.",
                }
                Tmpl.ExecuteTemplate(w, "edit_order.html", data)
                return
            }
            deadline = parsedDeadline
        } else {
            deadline = nil
        }

        _, err := db.DB.Exec("UPDATE orders SET order_name = ?, items = ?, deadline = ? WHERE id = ?", orderName, items, deadline, orderID)
        if err != nil {
            log.Printf("Error updating order ID %s in database: %v", orderID, err)
            data := PageData{
                Title:   "Edit Order",
                Year:    time.Now().Year(),
                OrderID: orderID,
                IsAdmin: isAdmin,
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Error updating order. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "edit_order.html", data)
            return
        }

        r.ParseMultipartForm(32 << 20) // 32 MB
        files := r.MultipartForm.File["images"]
        for _, fileHeader := range files {
            file, err := fileHeader.Open()
            if err != nil {
                log.Printf("Error opening image file for order ID %s: %v", orderID, err)
                data := PageData{
                    Title:   "Edit Order",
                    Year:    time.Now().Year(),
                    OrderID: orderID,
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "edit_order.html", data)
                return
            }
            defer file.Close()

            imageData, err := io.ReadAll(file)
            if err != nil {
                log.Printf("Error reading image file for order ID %s: %v", orderID, err)
                data := PageData{
                    Title:   "Edit Order",
                    Year:    time.Now().Year(),
                    OrderID: orderID,
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "edit_order.html", data)
                return
            }

            _, err = db.DB.Exec("INSERT INTO order_images (order_id, image_data) VALUES (?, ?)", orderID, imageData)
            if err != nil {
                log.Printf("Error inserting image into database for order ID %s: %v", orderID, err)
                data := PageData{
                    Title:   "Edit Order",
                    Year:    time.Now().Year(),
                    OrderID: orderID,
                    IsAdmin: isAdmin,
                    Theme:   theme,
                    LogoURL: logoURL,
                    Error:   "Error uploading image. Please try again.",
                }
                Tmpl.ExecuteTemplate(w, "edit_order.html", data)
                return
            }
        }

        log.Printf("Order ID %s updated successfully by user %s", orderID, userUUID.String())
        http.Redirect(w, r, "/view_order?id="+orderID, http.StatusSeeOther)
    }
}

func CloseOrderHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        log.Println("Close order failed: Invalid request method")
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    orderID := r.FormValue("orderID")
    if orderID == "" {
        log.Println("Close order failed: No order ID provided")
        http.Error(w, "Order ID is required", http.StatusBadRequest)
        return
    }

    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("Close order access denied: No userUUID cookie found")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("Close order access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    var userID int
    err = db.DB.QueryRow("SELECT id FROM users WHERE uuid = ?", userUUID.String()).Scan(&userID)
    if err != nil {
        log.Printf("Error fetching user ID for UUID %s during order close: %v", userUUID.String(), err)
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    _, err = db.DB.Exec("UPDATE orders SET closed = TRUE WHERE id = ?", orderID)
    if err != nil {
        log.Printf("Error closing order ID %s for user %s: %v", orderID, userUUID.String(), err)
        http.Error(w, "Error closing order", http.StatusInternalServerError)
        return
    }

    log.Printf("Order ID %s closed successfully by user %s", orderID, userUUID.String())
    http.Redirect(w, r, "/orders", http.StatusSeeOther)
}

func ServeImageHandler(w http.ResponseWriter, r *http.Request) {
    imageID := r.URL.Path[len("/image/"):]
    id, err := strconv.Atoi(imageID)
    if err != nil {
        log.Printf("Invalid image ID requested: %s, error: %v", imageID, err)
        http.Error(w, "Invalid image ID", http.StatusBadRequest)
        return
    }

    var imageData []byte
    err = db.DB.QueryRow("SELECT image_data FROM order_images WHERE id = ?", id).Scan(&imageData)
    if err != nil {
        log.Printf("Error fetching image ID %d: %v", id, err)
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