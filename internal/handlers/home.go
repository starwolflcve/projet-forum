package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"forum/internal/auth"
)

type App struct {
	db           *sql.DB
	templates    *template.Template
	sessions     map[string]string
	sessionMutex sync.RWMutex
}

type RegisterForm struct {
	FirstName string
	LastName  string
	Username  string
	Age       string
	Birthdate string
	Country   string
	Email     string
	Phone     string
	Gender    string
	Error     string
}

func New(db *sql.DB) (*App, error) {
	funcs := template.FuncMap{
		"eq": func(a, b string) bool { return a == b },
	}
	templates, err := template.New("app").Funcs(funcs).ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &App{
		db:        db,
		templates: templates,
		sessions:  make(map[string]string),
	}, nil
}

func (a *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.userFromRequest(r); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	a.renderTemplate(w, "index.html", nil)
}

func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.renderTemplate(w, "login.html", nil)
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Requête invalide", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			a.renderTemplate(w, "login.html", map[string]string{"Error": "Tous les champs sont obligatoires"})
			return
		}

		if !a.authenticateUser(username, password) {
			a.renderTemplate(w, "login.html", map[string]string{"Error": "Identifiant ou mot de passe incorrect"})
			return
		}

		token, err := auth.RandomToken()
		if err != nil {
			http.Error(w, "Erreur interne", http.StatusInternalServerError)
			return
		}

		a.storeSession(token, username)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func (a *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.renderTemplate(w, "register.html", nil)
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Requête invalide", http.StatusBadRequest)
			return
		}

		form := RegisterForm{
			FirstName: r.FormValue("first_name"),
			LastName:  r.FormValue("last_name"),
			Username:  r.FormValue("username"),
			Age:       r.FormValue("age"),
			Birthdate: r.FormValue("birthdate"),
			Country:   r.FormValue("country"),
			Email:     r.FormValue("email"),
			Phone:     r.FormValue("phone"),
			Gender:    r.FormValue("gender"),
		}
		password := r.FormValue("password")

		log.Printf("Registration attempt: %+v", form)

		if err := a.validateRegistration(&form, password); err != nil {
			log.Printf("Validation error: %v", err)
			form.Error = err.Error()
			a.renderTemplate(w, "register.html", form)
			return
		}

		age, _ := strconv.Atoi(form.Age)
		if err := a.insertUser(form, password, age); err != nil {
			log.Printf("Database insertion error: %v", err)
			if strings.Contains(err.Error(), "UNIQUE") {
				form.Error = "Ce nom d'utilisateur est déjà utilisé"
				a.renderTemplate(w, "register.html", form)
				return
			}
			http.Error(w, "Erreur interne", http.StatusInternalServerError)
			return
		}

		log.Printf("User %s registered successfully", form.Username)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func (a *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		a.deleteSession(cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := a.userFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	a.renderTemplate(w, "dashboard.html", map[string]string{"Username": username})
}

func (a *App) authenticateUser(username, password string) bool {
	var storedHash string
	err := a.db.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, username).Scan(&storedHash)
	if err != nil {
		return false
	}
	return storedHash == auth.HashPassword(password)
}

func (a *App) validateRegistration(data *RegisterForm, password string) error {
	if data.FirstName == "" || data.LastName == "" || data.Username == "" || password == "" || data.Age == "" || data.Birthdate == "" || data.Country == "" || data.Email == "" || data.Phone == "" || data.Gender == "" {
		return fmt.Errorf("Tous les champs sont obligatoires")
	}
	if _, err := strconv.Atoi(data.Age); err != nil {
		return fmt.Errorf("L'âge doit être un nombre valide")
	}
	if !strings.Contains(data.Email, "@") {
		return fmt.Errorf("L'adresse email n'est pas valide")
	}
	return nil
}

func (a *App) insertUser(data RegisterForm, password string, age int) error {
	_, err := a.db.Exec(`INSERT INTO users (username, password_hash, first_name, last_name, age, birthdate, country, email, phone, gender) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		data.Username,
		auth.HashPassword(password),
		data.FirstName,
		data.LastName,
		age,
		data.Birthdate,
		data.Country,
		data.Email,
		data.Phone,
		data.Gender,
	)
	return err
}

func (a *App) storeSession(token, username string) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.sessions[token] = username
}

func (a *App) getSessionUser(token string) (string, bool) {
	a.sessionMutex.RLock()
	defer a.sessionMutex.RUnlock()
	username, ok := a.sessions[token]
	return username, ok
}

func (a *App) deleteSession(token string) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	delete(a.sessions, token)
}

func (a *App) userFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", false
	}
	return a.getSessionUser(cookie.Value)
}

func (a *App) renderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, fmt.Sprintf("Erreur de rendu : %v", err), http.StatusInternalServerError)
	}
}
