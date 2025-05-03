package handlers

import (
    "management-system/db"
    "net/http"
    "time"
    "log"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        log.Println("Admin access denied: No userUUID cookie found")
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        log.Printf("Admin access denied: Invalid UUID format in cookie: %v", err)
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    var theme string
    err = db.DB.QueryRow("SELECT is_admin, theme FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin, &theme)
    if err != nil || !isAdmin {
        log.Printf("Admin access denied for UUID %s: Not admin or error fetching data: %v", userUUID.String(), err)
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    // Fetch the logo URL from the database if available
    var logoURL string
    var logoData []byte
    err = db.DB.QueryRow("SELECT image_data FROM site_settings WHERE setting_key = 'site_logo'").Scan(&logoData)
    if err == nil && len(logoData) > 0 {
        logoURL = "/logo"
    }

    if r.Method == "POST" {
        username := r.FormValue("username")
        password := r.FormValue("password")
        log.Printf("Admin creating new user: %s", username)

        // Check if username already exists
        var existingUsername string
        err = db.DB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUsername)
        if err == nil {
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Username already exists. Please choose a different one.",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            log.Printf("Error hashing password for new user %s: %v", username, err)
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Error processing registration",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        userUUID := uuid.New().String()
        _, err = db.DB.Exec("INSERT INTO users (username, password, uuid) VALUES (?, ?, ?)", username, string(hashedPassword), userUUID)
        if err != nil {
            log.Printf("Error inserting user %s into database: %v", username, err)
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Registration failed. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        log.Printf("New user %s created successfully by admin", username)
        http.Redirect(w, r, "/admin", http.StatusSeeOther)
        return
    }

    if r.Method == "DELETE" {
        userID := r.FormValue("userID")
        log.Printf("Admin attempting to delete user ID: %s", userID)

        // Check if the user to be deleted is an admin
        var isAdminUser bool
        err = db.DB.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdminUser)
        if err != nil {
            log.Printf("Error fetching user data for deletion (ID: %s): %v", userID, err)
            http.Error(w, "Error fetching user data", http.StatusInternalServerError)
            return
        }

        if isAdminUser {
            log.Printf("Admin attempted to delete another admin account (ID: %s)", userID)
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Cannot delete an admin account.",
            }
            rows, err := db.DB.Query("SELECT id, username, uuid FROM users")
            if err != nil {
                log.Printf("Error fetching users list after failed deletion: %v", err)
                http.Error(w, "Error fetching users", http.StatusInternalServerError)
                return
            }
            defer rows.Close()

            var users []struct {
                ID       int
                Username string
                UUID     string
            }

            for rows.Next() {
                var user struct {
                    ID       int
                    Username string
                    UUID     string
                }
                if err := rows.Scan(&user.ID, &user.Username, &user.UUID); err != nil {
                    log.Printf("Error scanning user data after failed deletion: %v", err)
                    http.Error(w, "Error scanning user data", http.StatusInternalServerError)
                    return
                }
                users = append(users, user)
            }

            data.Users = users
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        _, err = db.DB.Exec("DELETE FROM users WHERE id = ?", userID)
        if err != nil {
            log.Printf("Error deleting user ID %s from database: %v", userID, err)
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Theme:   theme,
                LogoURL: logoURL,
                Error:   "Error deleting user. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        log.Printf("User ID %s deleted successfully by admin", userID)
        http.Redirect(w, r, "/admin", http.StatusSeeOther)
        return
    }

    rows, err := db.DB.Query("SELECT id, username, uuid FROM users")
    if err != nil {
        log.Printf("Error fetching users list for admin panel: %v", err)
        http.Error(w, "Error fetching users", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []struct {
        ID       int
        Username string
        UUID     string
    }

    for rows.Next() {
        var user struct {
            ID       int
            Username string
            UUID     string
        }
        if err := rows.Scan(&user.ID, &user.Username, &user.UUID); err != nil {
            log.Printf("Error scanning user data for admin panel: %v", err)
            http.Error(w, "Error scanning user data", http.StatusInternalServerError)
            return
        }
        users = append(users, user)
    }

    data := PageData{
        Title:   "Admin Panel",
        Message: "Manage Users",
        Year:    time.Now().Year(),
        Users:   users,
        Theme:   theme,
        LogoURL: logoURL,
    }
    Tmpl.ExecuteTemplate(w, "admin.html", data)
}