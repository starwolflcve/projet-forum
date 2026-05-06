package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"forum/internal/database"
)

func ActivityHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userID := 1 // temporaire, à remplacer par l'utilisateur connecté

		activities, err := database.ListActivityByUserID(db, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération de l'activité", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Title":      "Mon activité",
			"Activities": activities,
		}

		err = tmpl.ExecuteTemplate(w, "activity.html", data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage de la page activité", http.StatusInternalServerError)
			return
		}
	}
}