package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"forum/internal/database"
	"forum/internal/middleware"
	"forum/internal/models"
)

func ActivityHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userID, err := middleware.GetUserIDFromSession(db, r)
		if err != nil {
			http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
			return
		}

		activityMap, err := database.ListActivityByUserID(db, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération de l'activité", http.StatusInternalServerError)
			return
		}

		// Aplatir la map en une slice d'activités
		var activities []models.ActivityItem
		if posts, ok := activityMap["posts"]; ok {
			activities = append(activities, posts...)
		}
		if reactions, ok := activityMap["reactions"]; ok {
			activities = append(activities, reactions...)
		}
		if comments, ok := activityMap["comments"]; ok {
			activities = append(activities, comments...)
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