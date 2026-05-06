package middleware

import (
	"database/sql"
	"net/http"
	"time"
)

func RequireRole(db *sql.DB, allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Non autorisé", http.StatusUnauthorized)
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
					http.Error(w, "Session invalide", http.StatusUnauthorized)
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
				http.Error(w, "Session expirée", http.StatusUnauthorized)
				return
			}

			var role string
			err = db.QueryRow(`
				SELECT role
				FROM users
				WHERE id = ?
			`, userID).Scan(&role)
			if err != nil {
				http.Error(w, "Utilisateur introuvable", http.StatusUnauthorized)
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					next(w, r)
					return
				}
			}

			http.Error(w, "Accès interdit", http.StatusForbidden)
		}
	}
}