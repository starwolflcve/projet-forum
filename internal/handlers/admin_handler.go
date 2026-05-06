package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"forum/internal/database"
)

func AdminCategoriesHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		categories, err := database.ListCategories(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Title":      "Gestion des catégories",
			"Categories": categories,
		}

		err = tmpl.ExecuteTemplate(w, "admin_categories.html", data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage des catégories", http.StatusInternalServerError)
			return
		}
	}
}

func CreateCategoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "Nom de catégorie requis", http.StatusBadRequest)
			return
		}

		err := database.CreateCategory(db, name)
		if err != nil {
			http.Error(w, "Erreur lors de la création de la catégorie", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
	}
}

func DeleteCategoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		categoryIDStr := r.FormValue("category_id")
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			http.Error(w, "ID de catégorie invalide", http.StatusBadRequest)
			return
		}

		err = database.DeleteCategoryByID(db, categoryID)
		if err != nil {
			http.Error(w, "Erreur lors de la suppression de la catégorie", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
	}
}