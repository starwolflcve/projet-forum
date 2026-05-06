package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/internal/database"
	"forum/internal/handlers"
	"forum/internal/middleware"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := database.Open("./db/forum.db")
	if err != nil {
		log.Fatalf("Erreur ouverture DB : %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Base de données inaccessible : %v", err)
	}

	if err := database.EnsureSchema(db); err != nil {
		log.Fatalf("Erreur création du schéma : %v", err)
	}
	if err := database.EnsureDefaultUser(db); err != nil {
		log.Fatalf("Erreur création de l'utilisateur par défaut : %v", err)
	}

	app, err := handlers.New(db)
	if err != nil {
		log.Fatalf("Erreur initialisation des handlers : %v", err)
	}

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", app.HomeHandler)
	mux.HandleFunc("/login", app.LoginHandler)
	mux.HandleFunc("/register", app.RegisterHandler)
	mux.HandleFunc("/logout", app.LogoutHandler)
	mux.HandleFunc("/dashboard", app.DashboardHandler)

	// Charger les templates pour P3
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))

	// Middlewares de protection pour P3
	requireLogin := middleware.RequireLogin(db)
	adminOnly := middleware.RequireRole(db, "admin")
	adminOrModerator := middleware.RequireRole(db, "admin", "moderator")

	// Advanced features : utilisateur connecté
	mux.HandleFunc("/activity", requireLogin(handlers.ActivityHandler(db, tmpl)))
	mux.HandleFunc("/notifications", requireLogin(handlers.NotificationsHandler(db, tmpl)))
	mux.HandleFunc("/notifications/read", requireLogin(handlers.MarkNotificationAsReadHandler(db)))

	// Dashboard admin central
	mux.HandleFunc("/admin", adminOnly(handlers.AdminDashboardHandler(tmpl)))

	// Admin uniquement : gestion catégories et utilisateurs
	mux.HandleFunc("/admin/categories", adminOnly(handlers.AdminCategoriesHandler(db, tmpl)))
	mux.HandleFunc("/admin/categories/create", adminOnly(handlers.CreateCategoryHandler(db)))
	mux.HandleFunc("/admin/categories/delete", adminOnly(handlers.DeleteCategoryHandler(db)))

	mux.HandleFunc("/admin/users", adminOnly(handlers.AdminUsersHandler(db, tmpl)))
	mux.HandleFunc("/admin/users/promote", adminOnly(handlers.PromoteUserHandler(db)))
	mux.HandleFunc("/admin/users/demote", adminOnly(handlers.DemoteModeratorHandler(db)))

	// Modération : accessible aux modérateurs et admins
	mux.HandleFunc("/moderation", adminOrModerator(handlers.ModerationDashboardHandler(db, tmpl)))
	mux.HandleFunc("/moderation/report", adminOrModerator(handlers.ReportContentHandler(db)))
	mux.HandleFunc("/moderation/approve", adminOrModerator(handlers.ApproveReportHandler(db)))
	mux.HandleFunc("/moderation/reject", adminOrModerator(handlers.RejectReportHandler(db)))
	mux.HandleFunc("/moderation/delete", adminOrModerator(handlers.DeleteReportedContentHandler(db)))

	log.Println("Serveur lancé sur http://localhost:8080")
	err = http.ListenAndServe(":8080", middleware.LoggingMiddleware(mux))
	if err != nil {
		log.Fatal(err)
	}
}
