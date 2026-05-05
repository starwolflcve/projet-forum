package main

import (
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

	log.Println("Serveur lancé sur http://localhost:8080")
	err = http.ListenAndServe(":8080", middleware.LoggingMiddleware(mux))
	if err != nil {
		log.Fatal(err)
	}
}
