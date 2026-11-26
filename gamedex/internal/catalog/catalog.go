package catalog

import (
	"fmt"
	"sort"

	"gamedex/internal/api"
)

type (
	Game       = api.Game
	Tournament = api.Tournament
)

type Catalog struct {
	client      *api.Client
	games       []Game
	genres      []string
	tournaments []Tournament
	byGame      map[int][]Tournament
}

func Load(client *api.Client) (*Catalog, error) {
	games, err := client.FetchGames()
	if err != nil {
		return nil, fmt.Errorf("récupération des jeux : %w", err)
	}
	tournaments, err := client.FetchTournaments()
	if err != nil {
		return nil, fmt.Errorf("récupération des tournois : %w", err)
	}

	byGame := make(map[int][]Tournament)
	for _, t := range tournaments {
		byGame[t.GameID] = append(byGame[t.GameID], t)
	}

	return &Catalog{
		client:      client,
		games:       games,
		genres:      genresOf(games),
		tournaments: tournaments,
		byGame:      byGame,
	}, nil
}

func genresOf(games []Game) []string {
	seen := make(map[string]bool)
	for _, g := range games {
		for _, genre := range g.Genres {
			seen[genre] = true
		}
	}
	genres := make([]string, 0, len(seen))
	for genre := range seen {
		genres = append(genres, genre)
	}
	sort.Strings(genres)
	return genres
}

func (c *Catalog) Games() []Game {
	return c.games
}

func (c *Catalog) Featured(n int) []Game {
	return c.games[:min(n, len(c.games))]
}

func (c *Catalog) Genres() []string {
	return c.genres
}

func (c *Catalog) Search(q, genre string) ([]Game, error) {
	return c.client.SearchGames(q, genre)
}

func (c *Catalog) Game(id int) (Game, bool) {
	for _, g := range c.games {
		if g.ID == id {
			return g, true
		}
	}
	return Game{}, false
}

func (c *Catalog) Tournaments(gameID int) []Tournament {
	return c.byGame[gameID]
}

func (c *Catalog) Tournament(id int) (Tournament, bool) {
	for _, t := range c.tournaments {
		if t.ID == id {
			return t, true
		}
	}
	return Tournament{}, false
}