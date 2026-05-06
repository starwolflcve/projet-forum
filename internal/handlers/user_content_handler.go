package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"forum/internal/database"
	"forum/internal/middleware"
)

func DeleteOwnPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userID, err := middleware.GetUserIDFromSession(db, r)
		if err != nil {
			http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
			return
		}

		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "ID de post invalide", http.StatusBadRequest)
			return
		}

		err = database.DeleteOwnPostByID(db, postID, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la suppression du post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/activity", http.StatusSeeOther)
	}
}

func DeleteOwnCommentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userID, err := middleware.GetUserIDFromSession(db, r)
		if err != nil {
			http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
			return
		}

		commentIDStr := r.FormValue("comment_id")
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
			return
		}

		err = database.DeleteOwnCommentByID(db, commentID, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la suppression du commentaire", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/activity", http.StatusSeeOther)
	}
}