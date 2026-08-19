package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lovelymondayz/say-less/backend/internal/matcher"
	"github.com/lovelymondayz/say-less/backend/internal/models"
	"github.com/lovelymondayz/say-less/backend/internal/spotify"
)

type Handler struct {
	container *models.Container
}

func NewHandler(c *models.Container) *Handler {
	return &Handler{container: c}
}

// New is an alias for NewHandler for backward compatibility
func New(c *models.Container) *Handler {
	return NewHandler(c)
}

type GenerateRequest struct {
	Text string `json:"text" binding:"required"`
	Mode string `json:"mode"`
}

type RegenerateRequest struct {
	OriginalText string `json:"original_text"`
	Mode         string `json:"mode"`
	ExcludeIDs   []string `json:"exclude_ids"`
}

type SearchTrackRequest struct {
	Query string `json:"query" binding:"required"`
}

type GenerateResponse struct {
	OriginalText   string        `json:"original_text"`
	NormalizedText string        `json:"normalized_text"`
	Mode           string        `json:"mode"`
	Tracks         []matcher.Track `json:"tracks"`
	Reconstructed  string        `json:"reconstructed"`
	Coverage       float64       `json:"coverage"`
	Caption        string        `json:"caption"`
}

type ShareRequest struct {
	OriginalText  string `json:"original_text"`
	Mode          string `json:"mode"`
	Tracks        []matcher.Track `json:"tracks"`
	Reconstructed string `json:"reconstructed"`
	Caption       string `json:"caption"`
	Coverage      float64 `json:"coverage"`
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := matcher.Mode(req.Mode)
	if mode == "" {
		mode = matcher.ModeSmart
	}

	resp := h.generateWithMode(req.Text, mode)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) generateWithMode(text string, mode matcher.Mode) GenerateResponse {
	normalized := matcher.Normalize(text)
	words := matcher.SplitWords(normalized)

	if len(words) == 0 {
		return GenerateResponse{
			OriginalText:   text,
			NormalizedText: normalized,
			Mode:           string(mode),
			Tracks:         []matcher.Track{},
			Reconstructed:  "",
			Coverage:       0,
			Caption:        "No matches found. Try a different phrase!",
		}
	}

	// Searcher function: searches Spotify for a phrase and returns all results
	searcher := func(phrase string) ([]matcher.Track, float64) {
		spotifyTracks, err := h.container.Spotify.SearchTracks(phrase, 20)
		if err != nil {
			log.Printf("Search error for '%s': %v", phrase, err)
			return nil, 0
		}
		log.Printf("Search '%s' returned %d results", phrase, len(spotifyTracks))

		var tracks []matcher.Track
		for _, st := range spotifyTracks {
			track := matcher.Track{
				ID:            st.ID,
				Title:         st.Name,
				Artist:        st.Artist,
				Album:         st.Album,
				ImageURL:      st.Image,
				PreviewURL:    st.Preview,
				Popularity:    st.Popularity,
				ArtistPopular: st.ArtistPopular,
				MatchScore:    scoreMatch(phrase, st, mode),
				MatchedPhrase: phrase,
			}

			// Determine match type
			phraseUpper := strings.ToUpper(phrase)
			trackUpper := strings.ToUpper(st.Name)
			if phraseUpper == trackUpper {
				track.MatchType = matcher.MatchExact
			} else if strings.Contains(trackUpper, phraseUpper) {
				track.MatchType = matcher.MatchPhrase
			} else if strings.Contains(phraseUpper, trackUpper) {
				track.MatchType = matcher.MatchPhrase
			} else {
				track.MatchType = matcher.MatchPartial
			}

			tracks = append(tracks, track)
		}

		return tracks, 0
	}

	// Use the matcher's DP to find best segmentation
	var tracks []matcher.Track
	if mode == matcher.ModeExact {
		tracks = matcher.FindOptimalExact(words, searcher, mode)
	} else {
		tracks = matcher.FindOptimalSegmentation(words, searcher, mode)
	}

	// If no results and input is Indonesian, try translated English
	if len(tracks) == 0 && matcher.IsIndonesian(text) {
		translated := matcher.TranslateIndonesian(text)
		if translated != normalized {
			transWords := matcher.SplitWords(translated)
			if mode == matcher.ModeExact {
				tracks = matcher.FindOptimalExact(transWords, searcher, mode)
			} else {
				tracks = matcher.FindOptimalSegmentation(transWords, searcher, mode)
			}
		}
	}

	if len(tracks) == 0 {
		return GenerateResponse{
			OriginalText:   text,
			NormalizedText: normalized,
			Mode:           string(mode),
			Tracks:         []matcher.Track{},
			Reconstructed:  "",
			Coverage:       0,
			Caption:        "No matches found. Try a different phrase!",
		}
	}

	var phrases []string
	for _, t := range tracks {
		phrases = append(phrases, t.MatchedPhrase)
	}
	reconstructed := strings.Join(phrases, " → ")
	coverage := calculateCoverage(words, tracks, normalized)
	caption := generateCaption()

	return GenerateResponse{
		OriginalText:   text,
		NormalizedText: normalized,
		Mode:           string(mode),
		Tracks:         tracks,
		Reconstructed:  reconstructed,
		Coverage:       coverage,
		Caption:        caption,
	}
}

func scoreMatch(phrase string, track spotify.Track, mode matcher.Mode) float64 {
	phraseUpper := strings.ToUpper(strings.TrimSpace(phrase))
	trackUpper := strings.ToUpper(strings.TrimSpace(track.Name))

	if phraseUpper == trackUpper {
		return 100.0
	}

	// Check if track title contains the phrase
	if strings.Contains(trackUpper, phraseUpper) {
		return 85.0 - float64(len(trackUpper)-len(phraseUpper))*0.5
	}

	// Check if phrase contains the track title
	if strings.Contains(phraseUpper, trackUpper) {
		return 75.0
	}

	// Word-level matching
	phraseWords := strings.Fields(phraseUpper)
	trackWords := strings.Fields(trackUpper)
	matchedWords := 0
	for _, pw := range phraseWords {
		for _, tw := range trackWords {
			if pw == tw {
				matchedWords++
				break
			}
		}
	}

	if len(phraseWords) == 0 {
		return 0
	}

	wordCoverage := float64(matchedWords) / float64(len(phraseWords))
	if mode == matcher.ModeChaos {
		return 30.0 + wordCoverage*20.0
	}

	if wordCoverage >= 0.7 {
		return 60.0 + wordCoverage*15.0
	} else if wordCoverage >= 0.4 {
		return 40.0 + wordCoverage*15.0
	}

	return wordCoverage * 30.0
}

func calculateCoverage(words []string, tracks []matcher.Track, normalized string) float64 {
	if len(words) == 0 {
		return 0
	}

	covered := make(map[string]bool)
	for _, t := range tracks {
		titleUpper := strings.ToUpper(t.Title)
		for _, w := range words {
			if strings.Contains(titleUpper, w) {
				covered[w] = true
			}
		}
	}

	return float64(len(covered)) / float64(len(words)) * 100.0
}

func generateCaption() string {
	captions := []string{
		"Your feelings are now a playlist.",
		"My music taste is now speaking for me.",
		"Spotify understood the assignment.",
		"Type it. We'll playlist it.",
		"I typed my thoughts and Spotify made them worse.",
	}
	return captions[time.Now().UnixNano()%int64(len(captions))]
}

func (h *Handler) Regenerate(c *gin.Context) {
	var req RegenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := matcher.Mode(req.Mode)
	if mode == "" {
		mode = matcher.ModeSmart
	}

	resp := h.generateWithMode(req.OriginalText, mode)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) SearchTrack(c *gin.Context) {
	var req SearchTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tracks, err := h.container.Spotify.SearchTracks(req.Query, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	var results []matcher.Track
	for _, st := range tracks {
		results = append(results, matcher.Track{
			ID:         st.ID,
			Title:      st.Name,
			Artist:     st.Artist,
			Album:      st.Album,
			ImageURL:   st.Image,
			PreviewURL: st.Preview,
			Popularity: st.Popularity,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tracks": results})
}

func (h *Handler) SaveShare(c *gin.Context) {
	var req ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tracksJSON, _ := json.Marshal(req.Tracks)
	shareID := spotify.RandomShareID()

	_, err := h.container.DB.Exec(context.Background(),
		`INSERT INTO generations (share_id, original_text, mode, tracks, reconstructed, caption, accuracy)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		shareID, req.OriginalText, req.Mode, tracksJSON, req.Reconstructed, req.Caption, req.Coverage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save share"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"share_id": shareID, "url": "/s/" + shareID})
}

func (h *Handler) GetShare(c *gin.Context) {
	shareID := c.Param("id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing share ID"})
		return
	}

	var originalText, mode, reconstructed, caption string
	var tracksJSON []byte
	var accuracy float64
	var createdAt time.Time
	err := h.container.DB.QueryRow(context.Background(),
		`SELECT original_text, mode, tracks, reconstructed, caption, accuracy, created_at
		 FROM generations WHERE share_id = $1`, shareID).
		Scan(&originalText, &mode, &tracksJSON, &reconstructed, &caption, &accuracy, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return
	}

	var tracks []matcher.Track
	json.Unmarshal(tracksJSON, &tracks)

	c.JSON(http.StatusOK, gin.H{
		"original_text":  originalText,
		"mode":           mode,
		"tracks":         tracks,
		"reconstructed":  reconstructed,
		"caption":        caption,
		"accuracy":       accuracy,
		"created_at":     createdAt,
	})
}

func (h *Handler) SpotifyLogin(c *gin.Context) {
	state := spotify.RandomShareID()
	url := h.container.Spotify.AuthURL(state)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *Handler) SpotifyCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}

	accessToken, refreshToken, err := h.container.Spotify.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

type CreatePlaylistRequest struct {
	Name     string   `json:"name" binding:"required"`
	TrackIDs []string `json:"track_ids" binding:"required"`
}

func (h *Handler) CreatePlaylist(c *gin.Context) {
	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.GetHeader("X-Spotify-Token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Spotify token"})
		return
	}

	result, err := h.container.Spotify.CreatePlaylist(token, req.Name, req.TrackIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Playlist creation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist_id":  result.ID,
		"playlist_url": result.URL,
		"name":         result.Name,
	})
}
