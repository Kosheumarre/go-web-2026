package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
	"gamedex/internal/server"
)

func main() {
	port := flag.String("port", "8080", "port d'écoute de GameDex")
	flag.Parse()

	cat, err := catalog.Load(api.NewClient(api.DefaultURL))
	if err != nil {
		log.Fatalf("GameDex ne peut pas démarrer, la mini API ne répond pas correctement (%s) : %v", api.DefaultURL, err)
	}
	afficherJeux(cat)

	srv, err := server.New(cat)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("GameDex sur http://localhost:%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, srv.Handler()))
}

func afficherJeux(cat *catalog.Catalog) {
	for _, g := range cat.Games() {
		fmt.Printf("%-25s (%d)  %d tournoi(s)\n", g.Name, g.Year, len(cat.Tournaments(g.ID)))
	}
}
