package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	clientID     string
	clientSecret string
	accessToken  string
	expiresAt    time.Time
	mu           sync.RWMutex
	httpClient   *http.Client
}

func NewClient() *Client {
	return &Client{
		clientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		clientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// EnsureToken refreshes the access token if needed
func (c *Client) EnsureToken() error {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("spotify token error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	c.accessToken = result.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn-30) * time.Second)

	return nil
}

type SearchResult struct {
	Tracks []Track `json:"tracks"`
}

type Track struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Artist  string   `json:"artist"`
	Album   string   `json:"album"`
	Image   string   `json:"image"`
	Preview string   `json:"preview_url"`
}

// SearchTracks searches for tracks matching a query
func (c *Client) SearchTracks(query string, limit int) ([]Track, error) {
	if err := c.EnsureToken(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=%d", encodedQuery, limit)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("spotify search error %d: %s", resp.StatusCode, string(body))
	}

	var rawResp struct {
		Tracks struct {
			Items []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Name   string `json:"name"`
					Images []struct {
						URL string `json:"url"`
					} `json:"images"`
				} `json:"album"`
				PreviewURL *string `json:"preview_url"`
			} `json:"items"`
		} `json:"tracks"`
	}

	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, err
	}

	var tracks []Track
	for _, item := range rawResp.Tracks.Items {
		track := Track{
			ID:     item.ID,
			Name:   item.Name,
			Artist: "",
			Album:  item.Album.Name,
			Preview: "",
		}
		if len(item.Artists) > 0 {
			track.Artist = item.Artists[0].Name
		}
		if len(item.Album.Images) > 0 {
			track.Image = item.Album.Images[0].URL
		}
		if item.PreviewURL != nil {
			track.Preview = *item.PreviewURL
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}
