package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"forum/internal/database"
	"forum/internal/models"
)

type ModerationPageData struct {
	Title   string
	Reports []models.Report
}

func ReportContentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		reporterID := 1 // temporaire, à remplacer par l'utilisateur connecté
		targetType := r.FormValue("target_type")
		targetIDStr := r.FormValue("target_id")
		reason := r.FormValue("reason")

		targetID, err := strconv.Atoi(targetIDStr)
		if err != nil {
			http.Error(w, "ID cible invalide", http.StatusBadRequest)
			return
		}

		if targetType != models.TargetTypePost && targetType != models.TargetTypeComment {
			http.Error(w, "Type de cible invalide", http.StatusBadRequest)
			return
		}

		switch reason {
		case models.ReportReasonIrrelevant, models.ReportReasonObscene, models.ReportReasonIllegal, models.ReportReasonInsulting:
		default:
			http.Error(w, "Raison de signalement invalide", http.StatusBadRequest)
			return
		}

		report := &models.Report{
			ReporterID: reporterID,
			TargetType: targetType,
			TargetID:   targetID,
			Reason:     reason,
			Status:     models.ReportStatusPending,
			CreatedAt:  time.Now(),
		}

		err = database.CreateReport(db, report)
		if err != nil {
			http.Error(w, "Erreur lors de la création du signalement", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func ModerationDashboardHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reports, err := database.ListPendingReports(db)
		if err != nil {
			http.Error(w, "Erreur lors du chargement des signalements", http.StatusInternalServerError)
			return
		}

		data := ModerationPageData{
			Title:   "Modération",
			Reports: reports,
		}

		err = tmpl.ExecuteTemplate(w, "moderation.html", data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage du dashboard", http.StatusInternalServerError)
			return
		}
	}
}