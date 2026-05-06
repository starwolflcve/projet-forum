package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"forum/internal/database"
)

func DeleteOwnPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		userID := 1 // temporaire, à remplacer par l'utilisateur connecté

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

		userID := 1 // temporaire, à remplacer par l'utilisateur connecté

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