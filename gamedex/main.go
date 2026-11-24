package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

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
  <p>Choisis une route à tester :</p>
  <ul class="menu">
    <li>
      <a href="/jeux">
        <span class="titre">Liste des jeux</span>
        <span class="route">GET /jeux</span>
        <span class="code ok">200</span>
      </a>
    </li>
    <li>
      <a href="/jeux/1">
        <span class="titre">Détail d'un jeu</span>
        <span class="route">GET /jeux/1</span>
        <span class="code ok">200</span>
      </a>
    </li>
    <li>
      <a href="/jeux/abc">
        <span class="titre">Identifiant invalide</span>
        <span class="route">GET /jeux/abc</span>
        <span class="code erreur">400</span>
      </a>
    </li>
    <li>
      <a href="/jeux/999">
        <span class="titre">Jeu introuvable</span>
        <span class="route">GET /jeux/999</span>
        <span class="code erreur">404</span>
      </a>
    </li>
    <li>
      <a href="/static/css/style.css">
        <span class="titre">Fichier statique</span>
        <span class="route">GET /static/css/style.css</span>
        <span class="code ok">200</span>
      </a>
    </li>
  </ul>
</body>
</html>`)
}

func listeJeux(w http.ResponseWriter, _ *http.Request) {
	games, err := FetchGames()
	if err != nil {
		log.Printf("FetchGames : %v", err)
		http.Error(w, "Service indisponible, réessayez plus tard.", http.StatusServiceUnavailable)
		return
	}
	for _, g := range games {
		fmt.Fprintf(w, "%d. %s (%d)\n", g.ID, g.Name, g.Year)
	}
}

func detailJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "l'identifiant doit être un nombre", http.StatusBadRequest)
		return
	}

	games, err := FetchGames()
	if err != nil {
		log.Printf("FetchGames : %v", err)
		http.Error(w, "Service indisponible, réessayez plus tard.", http.StatusServiceUnavailable)
		return
	}
	for _, g := range games {
		if g.ID == id {
			fmt.Fprintf(w, "%s (%d), studio %s\n", g.Name, g.Year, g.Studio)
			return
		}
	}
	http.Error(w, "jeu introuvable", http.StatusNotFound)
}