package catalog

import (
	"fmt"
	"strings"

	"gamedex/internal/api"
)

type (
	Game       = api.Game
	Tournament = api.Tournament
)

type Catalog struct {
	games  []Game
	byGame map[int][]Tournament
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
	return New(games, tournaments), nil
}

func New(games []Game, tournaments []Tournament) *Catalog {
	byGame := make(map[int][]Tournament)
	for _, t := range tournaments {
		byGame[t.GameID] = append(byGame[t.GameID], t)
	}
	return &Catalog{games: games, byGame: byGame}
}

func (c *Catalog) Games() []Game {
	return c.games
}

func (c *Catalog) Featured(n int) []Game {
	return c.games[:min(n, len(c.games))]
}

func (c *Catalog) Filter(q string) []Game {
	if q == "" {
		return c.games
	}
	q = strings.ToLower(q)
	result := []Game{}
	for _, g := range c.games {
		if strings.Contains(strings.ToLower(g.Name), q) {
			result = append(result, g)
		}
	}
	return result
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