package spotify

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

func isEnglish(text string) bool {
	for _, r := range text {
		if r > unicode.MaxASCII && !unicode.IsSpace(r) && !unicode.IsPunct(r) {
			return false
		}
	}
	return true
}

type Client struct {
	clientID     string
	clientSecret string
	accessToken  string
	expiresAt    time.Time
	mu           sync.RWMutex
	httpClient   *http.Client
	cache        map[string][]Track
	cacheMu      sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		clientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		clientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		cache:        make(map[string][]Track),
	}
}

func (c *Client) EnsureToken() error {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return nil
	}

	if c.clientID == "" || c.clientSecret == "" {
		return fmt.Errorf("spotify credentials not set")
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

type Track struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Artist        string `json:"artist"`
	ArtistPopular int    `json:"artist_popularity"`
	Album         string `json:"album"`
	Image         string `json:"image"`
	Preview       string `json:"preview_url"`
	Popularity    int    `json:"popularity"`
}

func (c *Client) SearchTracks(query string, limit int) ([]Track, error) {
	if err := c.EnsureToken(); err != nil {
		return nil, err
	}

	// Check cache first
	c.cacheMu.RLock()
	if cached, ok := c.cache[query]; ok {
		c.cacheMu.RUnlock()
		return cached, nil
	}
	c.cacheMu.RUnlock()

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

	// Retry with backoff on 429 - fail fast, return partial results
	var tracks []Track
	for attempt := 0; attempt < 2; attempt++ {
		req2, _ := http.NewRequest("GET", apiURL, nil)
		req2.Header.Set("Authorization", "Bearer "+token)
		resp2, err := c.httpClient.Do(req2)
		if err != nil {
			break // Don't retry on network errors
		}
		defer resp2.Body.Close()

		body, _ := io.ReadAll(resp2.Body)
		if resp2.StatusCode == 429 {
			// Rate limited - return empty (don't block)
			break
		}
		if resp2.StatusCode != 200 {
			return nil, fmt.Errorf("spotify search error %d: %s", resp2.StatusCode, string(body))
		}

		var rawResp struct {
			Tracks struct {
				Items []struct {
					ID         string `json:"id"`
					Name       string `json:"name"`
					Popularity int    `json:"popularity"`
					Artists    []struct {
						Name       string `json:"name"`
						Popularity int    `json:"popularity"`
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

		for _, item := range rawResp.Tracks.Items {
			track := Track{
				ID:         item.ID,
				Name:       item.Name,
				Popularity: item.Popularity,
			}
			if len(item.Artists) > 0 {
				track.Artist = item.Artists[0].Name
				track.ArtistPopular = item.Artists[0].Popularity
			}
			if item.Album.Name != "" {
				track.Album = item.Album.Name
			}
			if len(item.Album.Images) > 0 {
				track.Image = item.Album.Images[0].URL
			}
			if item.PreviewURL != nil {
				track.Preview = *item.PreviewURL
			}
			tracks = append(tracks, track)
		}
		break
	}

	// Cache results
	c.cacheMu.Lock()
	c.cache[query] = tracks
	c.cacheMu.Unlock()

	return tracks, nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (c *Client) AuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", os.Getenv("SPOTIFY_REDIRECT_URI"))
	params.Set("scope", "playlist-modify-public playlist-modify-private")
	params.Set("state", state)
	return "https://accounts.spotify.com/authorize?" + params.Encode()
}

func (c *Client) ExchangeCode(code string) (string, string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", os.Getenv("SPOTIFY_REDIRECT_URI"))

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("spotify exchange error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", err
	}

	return result.AccessToken, result.RefreshToken, nil
}

type PlaylistResult struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

func (c *Client) CreatePlaylist(token, name string, trackIDs []string) (*PlaylistResult, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
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
		return nil, fmt.Errorf("spotify me error %d: %s", resp.StatusCode, string(body))
	}

	var me struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &me); err != nil {
		return nil, err
	}

	createBody, _ := json.Marshal(map[string]interface{}{
		"name":   name,
		"public": true,
	})

	req, err = http.NewRequest("POST",
		fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", me.ID),
		strings.NewReader(string(createBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("spotify create playlist error %d: %s", resp.StatusCode, string(body))
	}

	var playlist struct {
		ID          string `json:"id"`
		ExternalURL struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &playlist); err != nil {
		return nil, err
	}

	if len(trackIDs) > 0 {
		uris := make([]string, len(trackIDs))
		for i, id := range trackIDs {
			uris[i] = "spotify:track:" + id
		}

		addBody, _ := json.Marshal(map[string]interface{}{
			"uris": uris,
		})

		req, err = http.NewRequest("POST",
			fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlist.ID),
			strings.NewReader(string(addBody)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
	}

	return &PlaylistResult{
		ID:   playlist.ID,
		URL:  playlist.ExternalURL.Spotify,
		Name: playlist.Name,
	}, nil
}

func RandomShareID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 10)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
