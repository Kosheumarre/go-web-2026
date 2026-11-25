package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

type PageListe struct {
	Games []Game
	Query string
}

type PageJeu struct {
	Game        Game
	Tournaments []Tournament
}

type PageAccueil struct {
	Games []Game
}

type PageErreur struct {
	Code    int
	Titre   string
	Message string
}

func main() {
	port := flag.String("port", "8080", "port d'écoute de GameDex")
	flag.Parse()

	if err := afficherJeux(); err != nil {
		log.Fatalf("GameDex ne peut pas démarrer, la mini API ne répond pas correctement (%s) : %v", apiURL, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", accueil)
	mux.HandleFunc("GET /jeux", listeJeux)
	mux.HandleFunc("GET /jeux/{id}", detailJeu)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", introuvable)

	log.Printf("GameDex sur http://localhost:%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}

func afficherJeux() error {
	games, err := FetchGames()
	if err != nil {
		return fmt.Errorf("récupération des jeux : %w", err)
	}
	tournaments, err := FetchTournaments()
	if err != nil {
		return fmt.Errorf("récupération des tournois : %w", err)
	}

	byGame := make(map[int][]Tournament)
	for _, t := range tournaments {
		byGame[t.GameID] = append(byGame[t.GameID], t)
	}

	for _, g := range games {
		fmt.Printf("%-25s (%d)  %d tournoi(s)\n", g.Name, g.Year, len(byGame[g.ID]))
	}
	return nil
}

func render(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("template %s : %v", name, err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func renderErreur(w http.ResponseWriter, code int, titre, message string) {
	render(w, code, "erreur.html", PageErreur{Code: code, Titre: titre, Message: message})
}

func erreurAPI(w http.ResponseWriter) {
	renderErreur(w, http.StatusServiceUnavailable, "Service indisponible",
		"Impossible de joindre la mini API. Réessayez dans un instant.")
}

func introuvable(w http.ResponseWriter, _ *http.Request) {
	renderErreur(w, http.StatusNotFound, "Page introuvable", "Cette adresse n'existe pas sur GameDex.")
}

func accueil(w http.ResponseWriter, _ *http.Request) {
	games, err := FetchGames()
	if err != nil {
		log.Printf("FetchGames : %v", err)
		erreurAPI(w)
		return
	}

	games = games[:min(4, len(games))]
	render(w, http.StatusOK, "accueil.html", PageAccueil{Games: games})
}

func listeJeux(w http.ResponseWriter, r *http.Request) {
	games, err := FetchGames()
	if err != nil {
		log.Printf("FetchGames : %v", err)
		erreurAPI(w)
		return
	}

	q := r.URL.Query().Get("q")
	if q != "" {
		filtres := []Game{}
		for _, g := range games {
			if strings.Contains(strings.ToLower(g.Name), strings.ToLower(q)) {
				filtres = append(filtres, g)
			}
		}
		games = filtres
	}

	render(w, http.StatusOK, "liste.html", PageListe{Games: games, Query: q})
}

func detailJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		renderErreur(w, http.StatusBadRequest, "Requête invalide", "L'identifiant d'un jeu doit être un nombre.")
		return
	}

	games, err := FetchGames()
	if err != nil {
		log.Printf("FetchGames : %v", err)
		erreurAPI(w)
		return
	}

	var jeu Game
	trouve := false
	for _, g := range games {
		if g.ID == id {
			jeu = g
			trouve = true
			break
		}
	}
	if !trouve {
		renderErreur(w, http.StatusNotFound, "Jeu introuvable", "Aucun jeu ne correspond à cet identifiant.")
		return
	}

	tournaments, err := FetchTournaments()
	if err != nil {
		log.Printf("FetchTournaments : %v", err)
		erreurAPI(w)
		return
	}
	duJeu := []Tournament{}
	for _, t := range tournaments {
		if t.GameID == id {
			duJeu = append(duJeu, t)
		}
	}

	render(w, http.StatusOK, "jeu.html", PageJeu{Game: jeu, Tournaments: duJeu})
}