package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"forum/internal/database"
	"forum/internal/middleware"
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

func ApproveReportHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		reviewedBy, err := middleware.GetUserIDFromSession(db, r)
		if err != nil {
			http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
			return
		}

		reportIDStr := r.FormValue("report_id")

		reportID, err := strconv.Atoi(reportIDStr)
		if err != nil {
			http.Error(w, "ID de signalement invalide", http.StatusBadRequest)
			return
		}

		err = database.UpdateReportStatus(db, reportID, models.ReportStatusReviewed, reviewedBy)
		if err != nil {
			http.Error(w, "Erreur lors de la validation du signalement", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/moderation", http.StatusSeeOther)
	}
}

func RejectReportHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		reviewedBy, err := middleware.GetUserIDFromSession(db, r)
		if err != nil {
			http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
			return
		}

		reportIDStr := r.FormValue("report_id")

		reportID, err := strconv.Atoi(reportIDStr)
		if err != nil {
			http.Error(w, "ID de signalement invalide", http.StatusBadRequest)
			return
		}

		err = database.UpdateReportStatus(db, reportID, models.ReportStatusRejected, reviewedBy)
		if err != nil {
			http.Error(w, "Erreur lors du rejet du signalement", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/moderation", http.StatusSeeOther)
	}
}

func DeleteReportedContentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		reportIDStr := r.FormValue("report_id")
		targetType := r.FormValue("target_type")
		targetIDStr := r.FormValue("target_id")

		reportID, err := strconv.Atoi(reportIDStr)
		if err != nil {
			http.Error(w, "ID de signalement invalide", http.StatusBadRequest)
			return
		}

		targetID, err := strconv.Atoi(targetIDStr)
		if err != nil {
			http.Error(w, "ID de cible invalide", http.StatusBadRequest)
			return
		}

		switch targetType {
		case "post":
			err = database.DeletePostByID(db, targetID)
		case "comment":
			err = database.DeleteCommentByID(db, targetID)
		default:
			http.Error(w, "Type de cible invalide", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Erreur lors de la suppression du contenu", http.StatusInternalServerError)
			return
		}

		reviewedBy, err := middleware.GetUserIDFromSession(db, r)
		if err != nil {
			http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
			return
		}

		err = database.UpdateReportStatus(db, reportID, models.ReportStatusReviewed, reviewedBy)
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour du signalement", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/moderation", http.StatusSeeOther)
	}
}