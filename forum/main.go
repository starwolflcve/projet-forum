package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/internal/database"
	"forum/internal/handlers"
	"forum/internal/middleware"
)

func main() {
	db, err := database.InitDB("db/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl := template.Must(template.ParseGlob("web/templates/pages/*.html"))

	// Middlewares de protection
	requireLogin := middleware.RequireLogin(db)
	adminOnly := middleware.RequireRole(db, "admin")
	adminOrModerator := middleware.RequireRole(db, "admin", "moderator")

	// Advanced features : utilisateur connecté
	http.HandleFunc("/activity", requireLogin(handlers.ActivityHandler(db, tmpl)))
	http.HandleFunc("/notifications", requireLogin(handlers.NotificationsHandler(db, tmpl)))
	http.HandleFunc("/notifications/read", requireLogin(handlers.MarkNotificationAsReadHandler(db)))

	// Dashboard admin central
	http.HandleFunc("/admin", adminOnly(handlers.AdminDashboardHandler(tmpl)))

	// Admin uniquement : gestion catégories et utilisateurs
	http.HandleFunc("/admin/categories", adminOnly(handlers.AdminCategoriesHandler(db, tmpl)))
	http.HandleFunc("/admin/categories/create", adminOnly(handlers.CreateCategoryHandler(db)))
	http.HandleFunc("/admin/categories/delete", adminOnly(handlers.DeleteCategoryHandler(db)))

	http.HandleFunc("/admin/users", adminOnly(handlers.AdminUsersHandler(db, tmpl)))
	http.HandleFunc("/admin/users/promote", adminOnly(handlers.PromoteUserHandler(db)))
	http.HandleFunc("/admin/users/demote", adminOnly(handlers.DemoteModeratorHandler(db)))

	// Modération : accessible aux modérateurs et admins
	http.HandleFunc("/moderation", adminOrModerator(handlers.ModerationDashboardHandler(db, tmpl)))
	http.HandleFunc("/moderation/report", adminOrModerator(handlers.ReportContentHandler(db)))
	http.HandleFunc("/moderation/approve", adminOrModerator(handlers.ApproveReportHandler(db)))
	http.HandleFunc("/moderation/reject", adminOrModerator(handlers.RejectReportHandler(db)))
	http.HandleFunc("/moderation/delete", adminOrModerator(handlers.DeleteReportedContentHandler(db)))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}