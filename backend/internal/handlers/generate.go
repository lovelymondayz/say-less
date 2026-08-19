package handlers

import (
	"context"
	"encoding/json"
	"log"
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
	OriginalText   string         `json:"original_text"`
	NormalizedText string         `json:"normalized_text"`
	Mode           string         `json:"mode"`
	Tracks         []matcher.Track `json:"tracks"`
	Reconstructed  string         `json:"reconstructed"`
	Coverage       float64        `json:"coverage"`
	Caption        string         `json:"caption"`
}

type PlaylistRequest struct {
	TrackIDs []string `json:"track_ids" binding:"required"`
	Token    string   `json:"token" binding:"required"`
	Name     string   `json:"name"`
}

type SearchTrackRequest struct {
	Phrase string `json:"phrase" binding:"required"`
	Mode   string `json:"mode"`
}

type RegenerateRequest struct {
	Text          string         `json:"text" binding:"required"`
	Mode          string         `json:"mode"`
	Strategy      string         `json:"strategy"`
	CurrentTracks []matcher.Track `json:"current_tracks"`
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

	result := h.generateWithMode(req.Text, mode)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) generateWithMode(text string, mode matcher.Mode) GenerateResponse {
	normalized := matcher.Normalize(text)
	words := matcher.SplitWords(normalized)

	searcher := func(phrase string) (*matcher.Track, float64) {
		return h.searchAndScore(phrase, mode)
	}

	tracks := matcher.FindOptimalSegmentation(words, searcher, mode)

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
	coverage := calculateCoverage(words, tracks)
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

	// If strategy changes mode, apply it
	if req.Strategy == "chaos" {
		mode = matcher.ModeChaos
	} else if req.Strategy == "accurate" || req.Strategy == "improve" {
		mode = matcher.ModeExact
	} else if req.Strategy == "popular" {
		mode = matcher.ModeSmart
	}

	// For "improve" and "popular", adjust scoring
	result := h.generateWithMode(req.Text, mode)

	c.JSON(http.StatusOK, result)
}

func (h *Handler) SearchTrack(c *gin.Context) {
	var req SearchTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := matcher.Mode(req.Mode)
	if mode == "" {
		mode = matcher.ModeSmart
	}

	track, score := h.searchAndScore(req.Phrase, mode)
	if track == nil {
		c.JSON(http.StatusOK, gin.H{"track": nil, "message": "No match found for this phrase"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"track": track, "score": score})
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
	tracks, err := h.container.Spotify.SearchTracks(phrase, 10)
	if err != nil {
		log.Printf("Spotify search error for '%s': %v", phrase, err)
		return nil, 0
	}
	if len(tracks) == 0 {
		return nil, 0
	}

	var bestTrack *matcher.Track
	bestScore := 0.0

	for _, t := range tracks {
		score := scoreMatch(phrase, t, mode)
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
				Popularity:    t.Popularity,
			}
		}
	}

	return bestTrack, bestScore
}

func scoreMatch(phrase string, track spotify.Track, mode matcher.Mode) float64 {
	phraseUpper := strings.ToUpper(strings.TrimSpace(phrase))
	trackUpper := strings.ToUpper(strings.TrimSpace(track.Name))

	baseScore := 0.0

	// Exact match
	if phraseUpper == trackUpper {
		baseScore = 100.0
	} else {
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
			baseScore = 95.0
		} else if strings.Contains(trackUpper, phraseUpper) {
			baseScore = 85.0 - float64(len(trackUpper)-len(phraseUpper))*0.5
		} else if strings.Contains(phraseUpper, trackUpper) {
			baseScore = 75.0
		} else {
			// Partial word match
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
					baseScore = 60.0 + wordCoverage*15.0
				} else if wordCoverage >= 0.4 {
					baseScore = 40.0 + wordCoverage*15.0
				} else {
					baseScore = 10.0
				}
			}
		}
	}

	// Popularity bonus (0-20 points based on track popularity 0-100)
	popularityBonus := float64(track.Popularity) * 0.2

	// Mode-specific adjustments
	switch mode {
	case matcher.ModeExact:
		// No popularity bonus, strict matching
		return baseScore
	case matcher.ModeSmart:
		// Slight popularity preference
		return baseScore + popularityBonus*0.3
	case matcher.ModeChaos:
		// More weight on popularity + some randomness for variety
		return baseScore + popularityBonus*0.5 + float64(time.Now().UnixNano()%10)*0.5
	}

	return baseScore
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

func generateCaption() string {
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
