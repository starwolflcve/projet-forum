package main

import (
  "database/sql"
  "log"
  "net/http"
  "path/filepath"

  "projet-forum/internal/database"
  "projet-forum/internal/handlers"

  _ "modernc.org/sqlite"
)

func main() {
  dbPath := filepath.Join("..", "db", "forum.db")
  db, err := sql.Open("sqlite", dbPath)
  if err != nil {
    log.Fatalf("failed to open database: %v", err)
  }
  defer db.Close()

  repo, err := database.NewPostRepository(db)
  if err != nil {
    log.Fatalf("failed to initialize repository: %v", err)
  }

  if err := repo.EnsureSchema(); err != nil {
    log.Fatalf("failed to ensure database schema: %v", err)
  }

  if err := repo.SeedCategories(); err != nil {
    log.Fatalf("failed to seed categories: %v", err)
  }

  handlers.SetRepository(repo)

  fs := http.FileServer(http.Dir(filepath.Join("..", "web", "static")))
  http.Handle("/static/", http.StripPrefix("/static/", fs))
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
      http.NotFound(w, r)
      return
    }
    http.ServeFile(w, r, filepath.Join("..", "web", "templates", "index.html"))
  })

  http.HandleFunc("/api/posts", handlers.HomeHandler)
  http.HandleFunc("/api/posts/detail", handlers.PostDetailHandler)
  http.HandleFunc("/api/posts/create", handlers.CreatePostHandler)
  http.HandleFunc("/api/posts/update", handlers.UpdatePostHandler)
  http.HandleFunc("/api/posts/delete", handlers.DeletePostHandler)
  http.HandleFunc("/api/posts/react", handlers.ReactToPostHandler)

  http.HandleFunc("/api/comments/create", handlers.CreateCommentHandler)
  http.HandleFunc("/api/comments/update", handlers.UpdateCommentHandler)
  http.HandleFunc("/api/comments/delete", handlers.DeleteCommentHandler)
  http.HandleFunc("/api/comments/react", handlers.ReactToCommentHandler)

  http.HandleFunc("/api/categories", handlers.ListCategoriesHandler)

  http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.Dir(filepath.Join("..", "web", "pages")))))

  log.Println("Starting forum server on http://localhost:8080")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
