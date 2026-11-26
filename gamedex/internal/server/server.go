package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"gamedex/internal/catalog"
	"gamedex/internal/store"
)

type PageAccueil struct {
	Games []catalog.Game
	Likes map[int]int
}

type PageListe struct {
	Games  []catalog.Game
	Genres []string
	Genre  string
	Query  string
	Likes  map[int]int
}

type TournoiVue struct {
	catalog.Tournament
	Inscrits []string
}

type PageJeu struct {
	Game        catalog.Game
	Tournaments []TournoiVue
	Likes       int
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
	mux.HandleFunc("POST /jeux/{id}/like", s.like)
	mux.HandleFunc("POST /tournois/{id}/inscription", s.inscription)
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

func (s *Server) findGame(w http.ResponseWriter, r *http.Request) (catalog.Game, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.renderErreur(w, http.StatusBadRequest, "Requête invalide", "L'identifiant d'un jeu doit être un nombre.")
		return catalog.Game{}, false
	}
	jeu, ok := s.catalog.Game(id)
	if !ok {
		s.renderErreur(w, http.StatusNotFound, "Jeu introuvable", "Aucun jeu ne correspond à cet identifiant.")
		return catalog.Game{}, false
	}
	return jeu, true
}

func (s *Server) introuvable(w http.ResponseWriter, _ *http.Request) {
	s.renderErreur(w, http.StatusNotFound, "Page introuvable", "Cette adresse n'existe pas sur GameDex.")
}

func trierParLikes(games []catalog.Game) map[int]int {
	likes := make(map[int]int, len(games))
	for _, g := range games {
		likes[g.ID] = store.Likes(g.ID)
	}
	sort.SliceStable(games, func(i, j int) bool {
		return likes[games[i].ID] > likes[games[j].ID]
	})
	return likes
}

func (s *Server) accueil(w http.ResponseWriter, _ *http.Request) {
	games := slices.Clone(s.catalog.Games())
	likes := trierParLikes(games)
	s.render(w, http.StatusOK, "accueil.html", PageAccueil{
		Games: games[:min(4, len(games))],
		Likes: likes,
	})
}

func (s *Server) listeJeux(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	genre := r.URL.Query().Get("genre")

	games, err := s.catalog.Search(q, genre)
	if err != nil {
		log.Printf("Search : %v", err)
		s.renderErreur(w, http.StatusBadGateway, "Service indisponible", "Le service des jeux ne répond pas.")
		return
	}

	likes := trierParLikes(games)

	s.render(w, http.StatusOK, "liste.html", PageListe{
		Games:  games,
		Genres: s.catalog.Genres(),
		Genre:  genre,
		Query:  q,
		Likes:  likes,
	})
}

func (s *Server) detailJeu(w http.ResponseWriter, r *http.Request) {
	jeu, ok := s.findGame(w, r)
	if !ok {
		return
	}

	tournois := []TournoiVue{}
	for _, t := range s.catalog.Tournaments(jeu.ID) {
		tournois = append(tournois, TournoiVue{Tournament: t, Inscrits: store.Registered(t.ID)})
	}

	s.render(w, http.StatusOK, "jeu.html", PageJeu{
		Game:        jeu,
		Tournaments: tournois,
		Likes:       store.Likes(jeu.ID),
	})
}

func (s *Server) like(w http.ResponseWriter, r *http.Request) {
	jeu, ok := s.findGame(w, r)
	if !ok {
		return
	}
	store.AddLike(jeu.ID)
	http.Redirect(w, r, "/jeux/"+strconv.Itoa(jeu.ID), http.StatusSeeOther)
}

func (s *Server) inscription(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.renderErreur(w, http.StatusBadRequest, "Requête invalide", "L'identifiant d'un tournoi doit être un nombre.")
		return
	}
	tournoi, ok := s.catalog.Tournament(id)
	if !ok {
		s.renderErreur(w, http.StatusNotFound, "Tournoi introuvable", "Aucun tournoi ne correspond à cet identifiant.")
		return
	}

	pseudo := strings.TrimSpace(r.PostFormValue("pseudo"))
	if pseudo == "" || utf8.RuneCountInString(pseudo) > 20 {
		s.renderErreur(w, http.StatusBadRequest, "Pseudo invalide", "Le pseudo doit faire entre 1 et 20 caractères.")
		return
	}

	store.Register(tournoi.ID, pseudo)
	http.Redirect(w, r, "/jeux/"+strconv.Itoa(tournoi.GameID), http.StatusSeeOther)
}