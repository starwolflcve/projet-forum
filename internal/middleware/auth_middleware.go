package middleware

import (
	"database/sql"
	"net/http"
	"time"
)

func RequireLogin(db *sql.DB) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			var userID int
			var expiresAt string

			err = db.QueryRow(`
				SELECT user_id, expires_at
				FROM sessions
				WHERE session_id = ?
			`, cookie.Value).Scan(&userID, &expiresAt)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				http.Error(w, "Erreur serveur", http.StatusInternalServerError)
				return
			}

			expirationTime, err := time.Parse("2006-01-02 15:04:05", expiresAt)
			if err != nil {
				http.Error(w, "Erreur format session", http.StatusInternalServerError)
				return
			}

			if time.Now().After(expirationTime) {
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					Expires:  time.Unix(0, 0),
					MaxAge:   -1,
				})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next(w, r)
		}
	}
}

func GetUserIDFromSession(db *sql.DB, r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		return 0, err
	}

	var userID int
	var expiresAt string

	err = db.QueryRow(`
		SELECT user_id, expires_at
		FROM sessions
		WHERE session_id = ?
	`, cookie.Value).Scan(&userID, &expiresAt)
	if err != nil {
		return 0, err
	}

	expirationTime, err := time.Parse("2006-01-02 15:04:05", expiresAt)
	if err != nil || time.Now().After(expirationTime) {
		return 0, err
	}

	return userID, nil
}