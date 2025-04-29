package handlers

import (
    "management-system/db"
    "net/http"
    "time"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "log"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("userUUID")
    if err != nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    userUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    var isAdmin bool
    err = db.DB.QueryRow("SELECT is_admin FROM users WHERE uuid = ?", userUUID.String()).Scan(&isAdmin)
    if err != nil || !isAdmin {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    if r.Method == "POST" {
        username := r.FormValue("username")
        password := r.FormValue("password")

        // Check if username already exists
        var existingUsername string
        err = db.DB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUsername)
        if err == nil {
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Error:   "Username already exists. Please choose a different one.",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            log.Printf("Error hashing password: %v", err)
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Error:   "Error processing registration",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        userUUID := uuid.New().String()
        _, err = db.DB.Exec("INSERT INTO users (username, password, uuid) VALUES (?, ?, ?)", username, string(hashedPassword), userUUID)
        if err != nil {
            log.Printf("Error inserting user into database: %v", err)
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Error:   "Registration failed. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        http.Redirect(w, r, "/admin", http.StatusSeeOther)
        return
    }

    if r.Method == "DELETE" {
        userID := r.FormValue("userID")

        // Check if the user to be deleted is the admin
        var isAdmin bool
        err = db.DB.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
        if err != nil {
            http.Error(w, "Error fetching user data", http.StatusInternalServerError)
            return
        }

        if isAdmin {
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Error:   "Cannot delete the admin account.",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        _, err = db.DB.Exec("DELETE FROM users WHERE id = ?", userID)
        if err != nil {
            log.Printf("Error deleting user from database: %v", err)
            data := PageData{
                Title:   "Admin Panel",
                Message: "Manage Users",
                Year:    time.Now().Year(),
                Error:   "Error deleting user. Please try again.",
            }
            Tmpl.ExecuteTemplate(w, "admin.html", data)
            return
        }

        http.Redirect(w, r, "/admin", http.StatusSeeOther)
        return
    }

    rows, err := db.DB.Query("SELECT id, username, uuid FROM users")
    if err != nil {
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
    }
    Tmpl.ExecuteTemplate(w, "admin.html", data)
}