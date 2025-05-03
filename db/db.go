package db

import (
    "database/sql"
    "log"

    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/crypto/bcrypt"
    "github.com/google/uuid"
)

var DB *sql.DB

func InitDB() {
    var err error
    DB, err = sql.Open("sqlite3", "./users.db")
    if err != nil {
        log.Fatal(err)
    }

    createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        uuid TEXT UNIQUE,
        is_admin BOOLEAN DEFAULT FALSE,
        theme TEXT DEFAULT 'light'
    );`
    _, err = DB.Exec(createUsersTable)
    if err != nil {
        log.Fatal(err)
    }

    createOrdersTable := `
    CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_name TEXT NOT NULL,
        items TEXT NOT NULL,
        user_id INTEGER NOT NULL,
        closed BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        deadline TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`
    _, err = DB.Exec(createOrdersTable)
    if err != nil {
        log.Fatal(err)
    }

    createOrderImagesTable := `
    CREATE TABLE IF NOT EXISTS order_images (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_id INTEGER NOT NULL,
        image_data BLOB NOT NULL,
        FOREIGN KEY (order_id) REFERENCES orders(id)
    );`
    _, err = DB.Exec(createOrderImagesTable)
    if err != nil {
        log.Fatal(err)
    }

    createSiteSettingsTable := `
    CREATE TABLE IF NOT EXISTS site_settings (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        setting_key TEXT UNIQUE NOT NULL,
        image_data BLOB
    );`
    _, err = DB.Exec(createSiteSettingsTable)
    if err != nil {
        log.Fatal(err)
    }

    // Check if admin user exists, if not, create one
    var adminUsername string
    err = DB.QueryRow("SELECT username FROM users WHERE username = 'admin'").Scan(&adminUsername)
    if err != nil {
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
        if err != nil {
            log.Fatal("Error hashing admin password:", err)
        }
        adminUUID := uuid.New().String()
        _, err = DB.Exec("INSERT INTO users (username, password, uuid, is_admin) VALUES (?, ?, ?, ?)", "admin", string(hashedPassword), adminUUID, true)
        if err != nil {
            log.Fatal("Error creating admin user:", err)
        }
    }
}