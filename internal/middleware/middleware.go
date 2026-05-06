package middleware

import (
	"database/sql"
	"net/http"
)

func RequireRole(db *sql.DB, allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Non autorisé", http.StatusUnauthorized)
				return
			}

			var userID int
			err = db.QueryRow(`
				SELECT user_id
				FROM sessions
				WHERE session_id = ?
			`, cookie.Value).Scan(&userID)
			if err != nil {
				http.Error(w, "Session invalide", http.StatusUnauthorized)
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