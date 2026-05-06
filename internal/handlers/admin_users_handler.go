package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"forum/internal/database"
	"forum/internal/models"
)

func AdminUsersHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		users, err := database.ListUsers(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Title": "Gestion des utilisateurs",
			"Users": users,
		}

		err = tmpl.ExecuteTemplate(w, "admin_users.html", data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage des utilisateurs", http.StatusInternalServerError)
			return
		}
	}
}

func PromoteUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.FormValue("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
			return
		}

		err = database.UpdateUserRole(db, userID, models.RoleModerator)
		if err != nil {
			http.Error(w, "Erreur lors de la promotion de l'utilisateur", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

func DemoteModeratorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.FormValue("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
			return
		}

		err = database.UpdateUserRole(db, userID, models.RoleUser)
		if err != nil {
			http.Error(w, "Erreur lors de la rétrogradation du modérateur", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}