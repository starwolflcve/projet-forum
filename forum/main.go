package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/internal/database"
	"forum/internal/handlers"
)

func main() {
	db, err := database.InitDB("db/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl := template.Must(template.ParseGlob("web/templates/pages/*.html"))

	http.HandleFunc("/activity", handlers.ActivityHandler(db, tmpl))
	http.HandleFunc("/notifications", handlers.NotificationsHandler(db, tmpl))
	http.HandleFunc("/notifications/read", handlers.MarkNotificationAsReadHandler(db))

	http.HandleFunc("/moderation", handlers.ModerationDashboardHandler(db, tmpl))
	http.HandleFunc("/moderation/report", handlers.ReportContentHandler(db))
	http.HandleFunc("/moderation/approve", handlers.ApproveReportHandler(db))
	http.HandleFunc("/moderation/reject", handlers.RejectReportHandler(db))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}