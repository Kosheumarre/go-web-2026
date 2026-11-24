package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func main() {
	port := flag.String("port", "8081", "port d'écoute")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", accueil)
	mux.HandleFunc("GET /jeux", listeJeux)
	mux.HandleFunc("GET /jeux/{id}", detailJeux)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Serveur sur http://localhost:8080")
	http.ListenAndServe(":8080", mux)

	log.Printf("Mini API sur http://localhost%s/api/", *port)
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func accueil(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <title>GameDex</title>
  <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
  <h1>Bienvenue sur GameDex</h1>
</body>
</html>`)
}

func listeJeux(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	genre := r.URL.Query().Get("genre")
	year := r.URL.Query().Get("year")

	result := []Game{}
	for _, g := range games {
		if q != "" && !strings.Contains(strings.ToLower(g.Name), q) {
			continue
		}
		if genre != "" && !slices.Contains(g.Genres, genre) {
			continue
		}
		if year != "" && strconv.Itoa(g.Year) != year {
			continue
		}
		result = append(result, g)
	}
	writeJSON(w, http.StatusOK, result)
}

func trouverJeux(w http.ResponseWriter, r *http.Request) (Game, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "l'identifiant doit être un nombre")
		return Game{}, false
	}
	for _, g := range games {
		if g.ID == id {
			return g, true
		}
	}
	writeError(w, http.StatusNotFound, "jeu introuvable")
	return Game{}, false
}

func detailJeux(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "l'identifiant doit être un nombre", http.StatusBadRequest)
		return
	}
	if id < 1 || id > 12 {
		http.Error(w, "jeu introuvable", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Jeu numéro %d", id)
}