package handlers

import (
	"context"
	"encoding/json"
	"math"
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

func New(c *models.Container) *Handler {
	return &Handler{container: c}
}

type GenerateRequest struct {
	Text string `json:"text" binding:"required"`
	Mode string `json:"mode"`
}

type GenerateResponse struct {
	OriginalText   string         `json:"original_text"`
	NormalizedText string         `json:"normalized_text"`
	Mode           string         `json:"mode"`
	Tracks         []matcher.Track `json:"tracks"`
	Reconstructed  string         `json:"reconstructed"`
	Coverage       float64        `json:"coverage"`
	Caption        string         `json:"caption"`
	ShareID        string         `json:"share_id,omitempty"`
}

type SaveShareRequest struct {
	OriginalText  string         `json:"original_text"`
	NormalizedText string        `json:"normalized_text"`
	Mode          string         `json:"mode"`
	Tracks        []matcher.Track `json:"tracks"`
	Reconstructed string         `json:"reconstructed"`
	Coverage      float64        `json:"coverage"`
	Caption       string         `json:"caption"`
}

type PlaylistRequest struct {
	TrackIDs []string `json:"track_ids" binding:"required"`
	Token    string   `json:"token" binding:"required"`
	Name     string   `json:"name"`
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

	normalized := matcher.Normalize(req.Text)
	words := matcher.SplitWords(normalized)

	searcher := func(phrase string) (*matcher.Track, float64) {
		return h.searchAndScore(phrase, mode)
	}

	tracks := matcher.FindOptimalSegmentation(words, searcher, mode)

	if len(tracks) == 0 {
		c.JSON(http.StatusOK, GenerateResponse{
			OriginalText:   req.Text,
			NormalizedText: normalized,
			Mode:           string(mode),
			Tracks:         []matcher.Track{},
			Reconstructed:  "",
			Coverage:       0,
			Caption:        "No matches found. Try a different phrase!",
		})
		return
	}

	var phrases []string
	for _, t := range tracks {
		phrases = append(phrases, t.MatchedPhrase)
	}
	reconstructed := strings.Join(phrases, " → ")
	coverage := calculateCoverage(words, tracks)
	caption := generateCaption(coverage)

	c.JSON(http.StatusOK, GenerateResponse{
		OriginalText:   req.Text,
		NormalizedText: normalized,
		Mode:           string(mode),
		Tracks:         tracks,
		Reconstructed:  reconstructed,
		Coverage:       coverage,
		Caption:        caption,
	})
}

func (h *Handler) SaveShare(c *gin.Context) {
	var req SaveShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shareID := spotify.RandomShareID()
	tracksJSON, _ := json.Marshal(req.Tracks)

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

	var gen models.Generation
	var tracksJSON []byte
	err := h.container.DB.QueryRow(context.Background(),
		`SELECT id, share_id, original_text, mode, tracks, reconstructed, caption, accuracy, created_at
		 FROM generations WHERE share_id = $1`, shareID).
		Scan(&gen.ID, &gen.ShareID, &gen.OriginalText, &gen.Mode, &tracksJSON,
			&gen.Reconstructed, &gen.Caption, &gen.Accuracy, &gen.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return
	}

	var tracks []matcher.Track
	json.Unmarshal(tracksJSON, &tracks)

	c.JSON(http.StatusOK, gin.H{
		"original_text":  gen.OriginalText,
		"mode":           gen.Mode,
		"tracks":         tracks,
		"reconstructed":  gen.Reconstructed,
		"caption":        gen.Caption,
		"accuracy":       gen.Accuracy,
		"created_at":     gen.CreatedAt,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) CreatePlaylist(c *gin.Context) {
	var req PlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		req.Name = "Say Less — My Playlist"
	}

	result, err := h.container.Spotify.CreatePlaylist(req.Token, req.Name, req.TrackIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) searchAndScore(phrase string, mode matcher.Mode) (*matcher.Track, float64) {
	tracks, err := h.container.Spotify.SearchTracks(phrase, 5)
	if err != nil || len(tracks) == 0 {
		return nil, 0
	}

	var bestTrack *matcher.Track
	bestScore := 0.0

	for _, t := range tracks {
		score := scoreMatch(phrase, t.Name, mode)
		if score > bestScore {
			bestScore = score
			mt := scoreToMatchType(score)
			bestTrack = &matcher.Track{
				ID:            t.ID,
				Title:         t.Name,
				Artist:        t.Artist,
				Album:         t.Album,
				ImageURL:      t.Image,
				PreviewURL:    t.Preview,
				MatchType:     mt,
				MatchScore:    score,
				MatchedPhrase: phrase,
			}
		}
	}

	return bestTrack, bestScore
}

func scoreMatch(phrase, trackTitle string, mode matcher.Mode) float64 {
	phraseUpper := strings.ToUpper(strings.TrimSpace(phrase))
	trackUpper := strings.ToUpper(strings.TrimSpace(trackTitle))

	if phraseUpper == trackUpper {
		return 100.0
	}

	stripPunct := func(s string) string {
		var b strings.Builder
		for _, r := range s {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
				b.WriteRune(r)
			}
		}
		return strings.Join(strings.Fields(b.String()), " ")
	}

	if stripPunct(phraseUpper) == stripPunct(trackUpper) {
		return 95.0
	}

	if strings.Contains(trackUpper, phraseUpper) {
		return 85.0 - float64(len(trackUpper)-len(phraseUpper))*0.5
	}

	if strings.Contains(phraseUpper, trackUpper) {
		return 75.0
	}

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

	if len(phraseWords) > 0 {
		wordCoverage := float64(matchedWords) / float64(len(phraseWords))
		if wordCoverage >= 0.7 {
			return 60.0 + wordCoverage*15.0
		}
		if wordCoverage >= 0.4 {
			return 40.0 + wordCoverage*15.0
		}
	}

	return 10.0
}

func scoreToMatchType(score float64) matcher.MatchType {
	switch {
	case score >= 95:
		return matcher.MatchExact
	case score >= 80:
		return matcher.MatchPhrase
	case score >= 60:
		return matcher.MatchPartial
	case score >= 40:
		return matcher.MatchSemantic
	default:
		return matcher.MatchWord
	}
}

func calculateCoverage(inputWords []string, tracks []matcher.Track) float64 {
	if len(inputWords) == 0 {
		return 0
	}
	covered := 0
	for _, t := range tracks {
		phraseWords := strings.Fields(strings.ToUpper(t.MatchedPhrase))
		covered += len(phraseWords)
	}
	ratio := float64(covered) / float64(len(inputWords))
	return math.Min(ratio, 1.0) * 100
}

func generateCaption(coverage float64) string {
	captions := []string{
		"Type it. We'll playlist it.",
		"Your thoughts, but make it music.",
		"Spotify understood the assignment.",
		"I typed my thoughts and Spotify made them worse.",
		"My music taste is now speaking for me.",
		"I said one thing. Spotify had another idea.",
		"Wait... these song titles actually say what I typed.",
		"Your feelings are now a playlist.",
	}
	idx := time.Now().UnixNano() % int64(len(captions))
	return captions[idx]
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
