package database

import (
    "database/sql"
    "fmt"

    "forum/internal/auth"
)

const defaultUsername = "admin"
const defaultPassword = "admin"

func Open(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }
    return db, nil
}

func EnsureSchema(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        first_name TEXT,
        last_name TEXT,
        age INTEGER,
        birthdate TEXT,
        country TEXT,
        email TEXT,
        phone TEXT,
        gender TEXT
    );`)
    if err != nil {
        return err
    }

    expected := map[string]string{
        "first_name":  "TEXT",
        "last_name":   "TEXT",
        "age":         "INTEGER",
        "birthdate":   "TEXT",
        "country":     "TEXT",
        "email":       "TEXT",
        "phone":       "TEXT",
        "gender":      "TEXT",
    }

    rows, err := db.Query(`PRAGMA table_info(users);`)
    if err != nil {
        return err
    }
    defer rows.Close()

    existing := map[string]bool{}
    for rows.Next() {
        var cid int
        var name, ctype string
        var notnull int
        var dfltValue sql.NullString
        var pk int
        if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
            return err
        }
        existing[name] = true
    }

    for column, ctype := range expected {
        if existing[column] {
            continue
        }
        _, err = db.Exec(fmt.Sprintf(`ALTER TABLE users ADD COLUMN %s %s;`, column, ctype))
        if err != nil {
            return err
        }
    }

    return nil
}

func EnsureDefaultUser(db *sql.DB) error {
    var count int
    err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, defaultUsername).Scan(&count)
    if err != nil {
        return err
    }
    if count > 0 {
        return nil
    }
    _, err = db.Exec(`INSERT INTO users (username, password_hash, first_name, last_name, age, birthdate, country, email, phone, gender) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        defaultUsername, auth.HashPassword(defaultPassword), "Admin", "Forum", 30, "1990-01-01", "France", "admin@example.com", "0000000000", "Autre")
    return err
}
