package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"gamedex/internal/catalog"
)

type PageAccueil struct {
	Games []catalog.Game
}

type PageListe struct {
	Games []catalog.Game
	Query string
}

type PageJeu struct {
	Game        catalog.Game
	Tournaments []catalog.Tournament
}

type PageErreur struct {
	Code    int
	Titre   string
	Message string
}

type Server struct {
	catalog *catalog.Catalog
	tmpl    *template.Template
}

func New(c *catalog.Catalog) (*Server, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("lecture des templates : %w", err)
	}
	return &Server{catalog: c, tmpl: tmpl}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.accueil)
	mux.HandleFunc("GET /jeux", s.listeJeux)
	mux.HandleFunc("GET /jeux/{id}", s.detailJeu)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", s.introuvable)
	return mux
}

func (s *Server) render(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("template %s : %v", name, err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) renderErreur(w http.ResponseWriter, code int, titre, message string) {
	s.render(w, code, "erreur.html", PageErreur{Code: code, Titre: titre, Message: message})
}

func (s *Server) introuvable(w http.ResponseWriter, _ *http.Request) {
	s.renderErreur(w, http.StatusNotFound, "Page introuvable", "Cette adresse n'existe pas sur GameDex.")
}

func (s *Server) accueil(w http.ResponseWriter, _ *http.Request) {
	s.render(w, http.StatusOK, "accueil.html", PageAccueil{Games: s.catalog.Featured(4)})
}

func (s *Server) listeJeux(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	s.render(w, http.StatusOK, "liste.html", PageListe{Games: s.catalog.Filter(q), Query: q})
}

func (s *Server) detailJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.renderErreur(w, http.StatusBadRequest, "Requête invalide", "L'identifiant d'un jeu doit être un nombre.")
		return
	}

	jeu, ok := s.catalog.Game(id)
	if !ok {
		s.renderErreur(w, http.StatusNotFound, "Jeu introuvable", "Aucun jeu ne correspond à cet identifiant.")
		return
	}

	s.render(w, http.StatusOK, "jeu.html", PageJeu{Game: jeu, Tournaments: s.catalog.Tournaments(id)})
}