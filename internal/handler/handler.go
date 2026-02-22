package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

// Project represents a portfolio project loaded from data/projects.json.
type Project struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Link        string   `json:"link"`
	Image       string   `json:"image"`
}

// Interest represents a personal interest loaded from data/interests.json.
type Interest struct {
	Emoji       string `json:"emoji"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// Experience represents a single entry in data/experience.json.
type Experience struct {
	Role        string   `json:"role"`
	Company     string   `json:"company"`
	CompanyURL  string   `json:"company_url"`
	Logo        string   `json:"logo"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	Description []string `json:"description"`
	Type        string   `json:"type"` // "work" or "education"
}

// About holds profile data loaded from data/about.json.
type About struct {
	Name              string `json:"name"`
	Tagline           string `json:"tagline"`
	Bio               string `json:"bio"`
	Location          string `json:"location"`
	Availability      bool   `json:"availability"`
	YearsOfExperience int    `json:"years_of_experience"`
	Email             string `json:"email"`
	GitHub            string `json:"github"`
	LinkedIn          string `json:"linkedin"`
	X                 string `json:"x"`
	ProfilePhoto      string `json:"profile_photo"`
}

// PageData is passed to all templates.
type PageData struct {
	About      About
	Projects   []Project
	Interests  []Interest
	Skills     []string
	Experience []Experience
}

// Handler holds parsed templates and pre-loaded page data.
type Handler struct {
	tmpl     *template.Template
	pageData PageData
}

// New creates a Handler by parsing templates and loading JSON data from fsys.
func New(fsys fs.FS) (*Handler, error) {
	tmpl, err := template.ParseFS(fsys, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	var about About
	if err := loadJSON(fsys, "data/about.json", &about); err != nil {
		return nil, fmt.Errorf("load about.json: %w", err)
	}

	var projects []Project
	if err := loadJSON(fsys, "data/projects.json", &projects); err != nil {
		return nil, fmt.Errorf("load projects.json: %w", err)
	}

	var interests []Interest
	if err := loadJSON(fsys, "data/interests.json", &interests); err != nil {
		return nil, fmt.Errorf("load interests.json: %w", err)
	}

	var skills []string
	if err := loadJSON(fsys, "data/skills.json", &skills); err != nil {
		return nil, fmt.Errorf("load skills.json: %w", err)
	}

	var experience []Experience
	if err := loadJSON(fsys, "data/experience.json", &experience); err != nil {
		return nil, fmt.Errorf("load experience.json: %w", err)
	}

	return &Handler{
		tmpl: tmpl,
		pageData: PageData{
			About:      about,
			Projects:   projects,
			Interests:  interests,
			Skills:     skills,
			Experience: experience,
		},
	}, nil
}

func loadJSON(fsys fs.FS, path string, v any) error {
	f, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func (h *Handler) execute(w http.ResponseWriter, name string, data any) {
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template %q error: %v", name, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// Index serves the full single-page application.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	h.execute(w, "base", h.pageData)
}

// About serves the about section partial for HTMX.
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	h.execute(w, "about", h.pageData)
}

// Projects serves the projects grid partial for HTMX.
func (h *Handler) Projects(w http.ResponseWriter, r *http.Request) {
	h.execute(w, "projects", h.pageData)
}

// Interests serves the interests grid partial for HTMX.
func (h *Handler) Interests(w http.ResponseWriter, r *http.Request) {
	h.execute(w, "interests", h.pageData)
}

// Contact handles the contact form POST and returns a success fragment.
func (h *Handler) Contact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	log.Printf("contact form submission: name=%q email=%q message_len=%d", name, email, len(message))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<div class="contact-success"><p>Thanks for reaching out — I'll be in touch soon.</p></div>`)
}

// Health returns 200 OK for health checks.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
