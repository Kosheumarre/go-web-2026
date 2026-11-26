package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)
const DefaultURL = "http://localhost:8081/api"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *Client) get(path string, dest any) error {
	resp, err := c.httpClient.Get(c.baseURL + path)
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

func (c *Client) FetchGames() ([]Game, error) {
	var games []Game
	if err := c.get("/games", &games); err != nil {
		return nil, err
	}
	return games, nil
}

func (c *Client) SearchGames(q, genre string) ([]Game, error) {
	params := url.Values{}
	params.Set("q", q)
	params.Set("genre", genre)

	var games []Game
	if err := c.get("/games?"+params.Encode(), &games); err != nil {
		return nil, err
	}
	return games, nil
}

func (c *Client) FetchTournaments() ([]Tournament, error) {
	var tournaments []Tournament
	if err := c.get("/tournaments", &tournaments); err != nil {
		return nil, err
	}
	return tournaments, nil
}