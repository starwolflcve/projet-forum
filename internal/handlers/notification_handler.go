package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"forum/internal/database"
)

func NotificationsHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userID := 1 // temporaire, à remplacer par l'utilisateur connecté

		notifications, err := database.ListNotificationsByUserID(db, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des notifications", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Title":         "Mes notifications",
			"Notifications": notifications,
		}

		err = tmpl.ExecuteTemplate(w, "notifications.html", data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage des notifications", http.StatusInternalServerError)
			return
		}
	}
}

func MarkNotificationAsReadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userID := 1 // temporaire, à remplacer par l'utilisateur connecté

		notificationIDStr := r.FormValue("notification_id")
		notificationID, err := strconv.Atoi(notificationIDStr)
		if err != nil {
			http.Error(w, "ID de notification invalide", http.StatusBadRequest)
			return
		}

		err = database.MarkNotificationAsRead(db, notificationID, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour de la notification", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/notifications", http.StatusSeeOther)
	}
}