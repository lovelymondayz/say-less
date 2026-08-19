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

	var selectedTitles []string

	searcher := func(phrase string) (*matcher.Track, float64) {
		return h.searchAndScore(phrase, mode, selectedTitles)
	}

	var tracks []matcher.Track
	if mode == matcher.ModeExact {
		tracks = matcher.FindOptimalExact(words, searcher, mode)
	} else {
		tracks = matcher.FindOptimalSegmentation(words, searcher, mode)
	}

	// Update selected titles after segmentation
	for _, t := range tracks {
		if t.Title != "[unavailable]" {
			selectedTitles = append(selectedTitles, strings.ToUpper(t.Title))
		}
	}

	// If no results and input is Indonesian, translate to English and retry
	if len(tracks) == 0 && matcher.IsIndonesian(text) {
		translated := matcher.TranslateIndonesian(text)
		if translated != normalized {
			engWords := matcher.SplitWords(translated)
			var engSelectedTitles []string
			engSearcher := func(phrase string) (*matcher.Track, float64) {
				return h.searchAndScore(phrase, mode, engSelectedTitles)
			}
			if mode == matcher.ModeExact {
				tracks = matcher.FindOptimalExact(engWords, engSearcher, mode)
			} else {
				tracks = matcher.FindOptimalSegmentation(engWords, engSearcher, mode)
			}
			for _, t := range tracks {
				if t.Title != "[unavailable]" {
					engSelectedTitles = append(engSelectedTitles, strings.ToUpper(t.Title))
				}
			}
		}
	}

	// Semantic fallback: if all tracks are single-word matches, try semantic for whole phrase
	if len(tracks) > 0 && len(tracks) == len(words) && mode == matcher.ModeSmart {
		semanticTrack, semanticScore := h.semanticFallback(text, mode, selectedTitles)
		if semanticTrack != nil && semanticScore > 50 {
			// Replace the word-by-word result with the semantic match
			tracks = []matcher.Track{*semanticTrack}
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

	if req.Strategy == "chaos" {
		mode = matcher.ModeChaos
	} else if req.Strategy == "accurate" || req.Strategy == "improve" {
		mode = matcher.ModeExact
	} else if req.Strategy == "popular" {
		mode = matcher.ModeSmart
	}

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

	track, score := h.searchAndScore(req.Phrase, mode, nil)
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

// searchAndScore searches Spotify for the phrase and scores the best track.
// It uses phrase containment filtering — only tracks whose title contains the phrase are accepted.
func (h *Handler) searchAndScore(phrase string, mode matcher.Mode, selectedTitles []string) (*matcher.Track, float64) {
	tracks, err := h.container.Spotify.SearchTracks(phrase, 10)
	if err != nil {
		log.Printf("Spotify search error for '%s': %v", phrase, err)
		return nil, 0
	}
	if len(tracks) == 0 {
		return h.semanticFallback(phrase, mode, selectedTitles)
	}

	var bestTrack *matcher.Track
	bestScore := 0.0

	for _, t := range tracks {
		score := scoreMatch(phrase, t, mode, selectedTitles)
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
				ArtistPopular: t.ArtistPopular,
			}
		}
	}

	return bestTrack, bestScore
}

// semanticFallback tries to find tracks matching semantic concepts related to the phrase.
func (h *Handler) semanticFallback(phrase string, mode matcher.Mode, selectedTitles []string) (*matcher.Track, float64) {
	keywords := matcher.FindSemanticKeywords(phrase)
	if len(keywords) == 0 {
		return nil, 0
	}

	var bestTrack *matcher.Track
	bestScore := 0.0

	limit := 5
	if len(keywords) < 5 {
		limit = len(keywords)
	}

	for i := 0; i < limit; i++ {
		kw := keywords[i]
		tracks, err := h.container.Spotify.SearchTracks(kw, 5)
		if err != nil || len(tracks) == 0 {
			continue
		}

		for _, t := range tracks {
			semanticScore := 45.0 + float64(t.Popularity)*0.2

			if isRepetitiveTitle(t.Name, selectedTitles) {
				semanticScore -= 30
			}

			if semanticScore > bestScore {
				bestScore = semanticScore
				bestTrack = &matcher.Track{
					ID:            t.ID,
					Title:         t.Name,
					Artist:        t.Artist,
					Album:         t.Album,
					ImageURL:      t.Image,
					PreviewURL:    t.Preview,
					MatchType:     matcher.MatchSemantic,
					MatchScore:    semanticScore,
					MatchedPhrase: phrase,
					Popularity:    t.Popularity,
					ArtistPopular: t.ArtistPopular,
				}
			}
		}
	}

	return bestTrack, bestScore
}

// isRepetitiveTitle checks if the title is similar to already-selected titles.
func isRepetitiveTitle(title string, selectedTitles []string) bool {
	titleUpper := strings.ToUpper(title)
	titleWords := strings.Fields(titleUpper)
	if len(titleWords) < 2 {
		return false
	}

	for _, sel := range selectedTitles {
		selWords := strings.Fields(sel)
		matchCount := 0
		for i := 0; i < len(titleWords) && i < len(selWords) && i < 3; i++ {
			if titleWords[i] == selWords[i] {
				matchCount++
			}
		}
		if matchCount >= 2 {
			return true
		}
	}
	return false
}

func scoreMatch(phrase string, track spotify.Track, mode matcher.Mode, selectedTitles []string) float64 {
	phraseUpper := strings.ToUpper(strings.TrimSpace(phrase))
	trackUpper := strings.ToUpper(strings.TrimSpace(track.Name))

	// Phrase containment filter: track title must contain the phrase as words
	phraseWords := strings.Fields(phraseUpper)
	trackWords := strings.Fields(trackUpper)

	// For single-word phrases, require exact word match (not substring)
	if len(phraseWords) == 1 {
		word := phraseWords[0]
		found := false
		for _, tw := range trackWords {
			if tw == word {
				found = true
				break
			}
		}
		if !found {
			return 0 // Reject — track doesn't contain the word
		}
	}

	// For multi-word phrases, check if track title contains all phrase words in order
	if len(phraseWords) > 1 {
		// Check if track title contains the full phrase
		containsPhrase := strings.Contains(trackUpper, phraseUpper)
		// Check if phrase contains the track title
		phraseContains := strings.Contains(phraseUpper, trackUpper)

		if !containsPhrase && !phraseContains {
			// Check if most words match in order
			matchedWords := 0
			trackIdx := 0
			for _, pw := range phraseWords {
				for trackIdx < len(trackWords) {
					if trackWords[trackIdx] == pw {
						matchedWords++
						trackIdx++
						break
					}
					trackIdx++
				}
			}
			if matchedWords < len(phraseWords)/2+1 {
				return 0 // Reject — track doesn't contain enough of the phrase in order
			}
		}
	}

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

	// Word order similarity bonus (LCS)
	lcsLength := lcs(phraseWords, trackWords)
	if lcsLength >= 2 && len(phraseWords) > 0 {
		orderBonus := (float64(lcsLength) / float64(len(phraseWords))) * 10.0
		baseScore += orderBonus
	}

	// Repetitive title avoidance
	if isRepetitiveTitle(track.Name, selectedTitles) {
		baseScore -= 30.0
	}

	// Artist recognizability boost (0-5 points)
	artistBonus := float64(track.ArtistPopular) * 0.05

	// Popularity bonus (0-20 points based on track popularity 0-100)
	popularityBonus := float64(track.Popularity) * 0.2

	// Mode-specific adjustments
	switch mode {
	case matcher.ModeExact:
		// No popularity bonus, strict matching
		return baseScore + artistBonus*0.5
	case matcher.ModeSmart:
		// Slight popularity preference, but accuracy matters more
		accuracyWeight := baseScore / 100.0
		if accuracyWeight < 0.5 {
			// Poor match — don't boost with popularity
			return baseScore + artistBonus*0.3
		}
		return baseScore + popularityBonus*0.2 + artistBonus*0.3
	case matcher.ModeChaos:
		// More weight on popularity + some randomness for variety
		return baseScore + popularityBonus*0.5 + artistBonus + float64(time.Now().UnixNano()%10)*0.5
	}

	return baseScore
}

// lcs computes the length of the longest common subsequence between two string slices.
func lcs(a, b []string) int {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return 0
	}

	prev := make([]int, n+1)
	curr := make([]int, n+1)

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				if curr[j-1] > prev[j] {
					curr[j] = curr[j-1]
				} else {
					curr[j] = prev[j]
				}
			}
		}
		prev, curr = curr, prev
		for k := range curr {
			curr[k] = 0
		}
	}

	return prev[n]
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

func calculateCoverage(inputWords []string, tracks []matcher.Track, originalText string) float64 {
	if len(inputWords) == 0 {
		return 0
	}
	covered := 0
	for _, t := range tracks {
		phraseWords := strings.Fields(strings.ToUpper(t.MatchedPhrase))
		covered += len(phraseWords)
	}
	ratio := float64(covered) / float64(len(inputWords))
	coverage := math.Min(ratio, 1.0) * 100

	// Semantic reconstruction check: if key emotional words are missing, reduce coverage
	if originalText != "" && len(tracks) > 0 {
		keyWords := extractKeyWords(originalText)
		if len(keyWords) > 0 {
			allTrackText := strings.Builder{}
			for _, t := range tracks {
				allTrackText.WriteString(strings.ToUpper(t.Title))
				allTrackText.WriteString(" ")
			}
			trackText := allTrackText.String()

			missingCount := 0
			for _, kw := range keyWords {
				if !strings.Contains(trackText, kw) {
					missingCount++
				}
			}

			if missingCount > 0 {
				missingRatio := float64(missingCount) / float64(len(keyWords))
				if missingRatio > 0.5 {
					coverage *= 0.8
				}
			}
		}
	}

	return coverage
}

// extractKeyWords extracts important emotional/content words (not stop words).
func extractKeyWords(text string) []string {
	stopWords := map[string]bool{
		"THE": true, "A": true, "AN": true, "AND": true, "OR": true,
		"BUT": true, "IN": true, "ON": true, "AT": true, "TO": true,
		"FOR": true, "OF": true, "WITH": true, "BY": true, "IS": true,
		"IT": true, "I": true, "YOU": true, "HE": true, "SHE": true,
		"WE": true, "THEY": true, "ME": true, "HIM": true, "HER": true,
		"US": true, "THEM": true, "MY": true, "YOUR": true, "HIS": true,
		"AM": true, "ARE": true, "WAS": true, "WERE": true, "BE": true,
		"DO": true, "DOES": true, "DID": true, "HAVE": true, "HAS": true,
		"HAD": true, "WILL": true, "WOULD": true, "COULD": true, "SHOULD": true,
		"SO": true, "NOT": true, "NO": true, "IF": true, "THAT": true,
		"THIS": true, "WHAT": true, "WHICH": true, "WHO": true, "WHEN": true,
		"HOW": true, "WHY": true, "ALL": true, "CAN": true,
	}

	words := strings.Fields(strings.ToUpper(text))
	var keyWords []string
	seen := make(map[string]bool)
	for _, w := range words {
		if !stopWords[w] && !seen[w] && len(w) > 1 {
			keyWords = append(keyWords, w)
			seen[w] = true
		}
	}
	return keyWords
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
