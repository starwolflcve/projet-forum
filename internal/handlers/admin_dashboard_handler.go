package handlers

import (
	"html/template"
	"net/http"
)

func AdminDashboardHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		data := map[string]interface{}{
			"Title": "Dashboard administrateur",
		}

		err := tmpl.ExecuteTemplate(w, "admin_dashboard.html", data)
		if err != nil {
			http.Error(w, "Erreur lors de l'affichage du dashboard admin", http.StatusInternalServerError)
			return
		}
	}
}