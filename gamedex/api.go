package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiURL = "http://localhost:8081/api"

var client = &http.Client{Timeout: 3 * time.Second}

type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}
type Tournament struct {
	ID     int    `json:"id"`
	GameID int    `json:"gameId"`
	Name   string `json:"name"`
	City   string `json:"city"`
	Date   string `json:"date"`
}

func fetchJSON(path string, dest any) error {
	resp, err := client.Get(apiURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("l'API a répondu %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("JSON illisible : %w", err)
	}
	return nil
}

func FetchGames() ([]Game, error) {
	var games []Game
	if err := fetchJSON("/games", &games); err != nil {
		return nil, err
	}
	return games, nil
}

func FetchTournaments() ([]Tournament, error) {
	var tournaments []Tournament
	if err := fetchJSON("/tournaments", &tournaments); err != nil {
		return nil, err
	}
	return tournaments, nil
}