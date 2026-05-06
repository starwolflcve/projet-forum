package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"forum/internal/database"
	"forum/internal/models"
)

type ActivityPageData struct {
	Title     string
	UserID    int
	Posts     []models.ActivityItem
	Reactions []models.ActivityItem
	Comments  []models.ActivityItem
}

func ActivityHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := 1 // temporaire, à remplacer plus tard par l'utilisateur connecté

		activityData, err := database.ListActivityByUserID(db, userID)
		if err != nil {
			http.Error(w, "Erreur lors du chargement de l'activité", http.StatusInternalServerError)
			return
		}

		data := ActivityPageData{
			Title:     "Mon activité",
			UserID:    userID,
			Posts:     activityData["posts"],
			Reactions: activityData["reactions"],
			Comments:  activityData["comments"],
		}

		err = tmpl.ExecuteTemplate(w, "activity.html", data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage de la page", http.StatusInternalServerError)
			return
		}
	}
}