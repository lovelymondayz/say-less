package matcher

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

type Mode string

const (
	ModeExact Mode = "exact"
	ModeSmart Mode = "smart"
	ModeChaos Mode = "chaos"
)

type MatchType string

const (
	MatchExact    MatchType = "exact"
	MatchPhrase   MatchType = "phrase"
	MatchPartial  MatchType = "partial"
	MatchSemantic MatchType = "semantic"
	MatchWord     MatchType = "word"
)

type Track struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Artist        string    `json:"artist"`
	Album         string    `json:"album"`
	ImageURL      string    `json:"image_url"`
	PreviewURL    string    `json:"preview_url,omitempty"`
	MatchType     MatchType `json:"match_type"`
	MatchScore    float64   `json:"match_score"`
	MatchedPhrase string    `json:"matched_phrase"`
	Popularity    int       `json:"popularity"`
	ArtistPopular int       `json:"artist_popularity"`
}

type Result struct {
	OriginalText   string  `json:"original_text"`
	NormalizedText string  `json:"normalized_text"`
	Mode           Mode    `json:"mode"`
	Tracks         []Track `json:"tracks"`
	Reconstructed  string  `json:"reconstructed"`
	Coverage       float64 `json:"coverage"`
	Caption        string  `json:"caption"`
}

func Normalize(input string) string {
	s := strings.ToUpper(input)

	// Replace common contractions
	replacements := map[string]string{
		"DON'T": "DO NOT", "WON'T": "WILL NOT", "CAN'T": "CANNOT",
		"ISN'T": "IS NOT", "AREN'T": "ARE NOT", "WASN'T": "WAS NOT",
		"WEREN'T": "WERE NOT", "HAVEN'T": "HAVE NOT", "HASN'T": "HAS NOT",
		"HADN'T": "HAD NOT", "DOESN'T": "DOES NOT", "DIDN'T": "DID NOT",
		"COULDN'T": "COULD NOT", "SHOULDN'T": "SHOULD NOT", "WOULDN'T": "WOULD NOT",
		"I'M": "I AM", "I'VE": "I HAVE", "I'LL": "I WILL", "I'D": "I WOULD",
		"IT'S": "IT IS", "THAT'S": "THAT IS", "WHAT'S": "WHAT IS", "WHO'S": "WHO IS",
		"LET'S": "LET US",
	}

	for contraction, expanded := range replacements {
		s = strings.ReplaceAll(s, contraction, expanded)
	}

	// Keep letters, numbers, spaces, emojis, and some meaningful chars
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '&' || r == '!' || r == '?' {
			b.WriteRune(r)
		} else if r == '\'' {
			continue
		} else if r > 127 {
			// Keep emojis and other non-ASCII
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}

	result := strings.Join(strings.Fields(b.String()), " ")
	return strings.TrimSpace(result)
}

func SplitWords(normalized string) []string {
	return strings.Fields(normalized)
}

func Phrase(words []string, start, end int) string {
	return strings.Join(words[start:end], " ")
}

// isEmoji checks if a string is purely emojis
func isEmoji(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < 0x1F600 && r < 0x2600 {
			return false
		}
	}
	return true
}

// shouldSkipWord returns true if the word should be skipped in phrase search
func shouldSkipWord(word string) bool {
	// Skip single-letter words (I, A)
	if len(word) <= 1 {
		return true
	}
	// Skip pure emojis
	if isEmoji(word) {
		return true
	}
	return false
}

type DPScore struct {
	Score       float64
	Phrase      string
	Track       *Track
	PrevIdx     int
	Unavailable bool
}

// FindOptimalSegmentation finds the best phrase-first segmentation.
// It uses dynamic programming with phrase containment filtering.
func FindOptimalSegmentation(words []string, searcher func(string) (*Track, float64), mode Mode) []Track {
	n := len(words)
	if n == 0 {
		return nil
	}

	dp := make([]DPScore, n+1)
	dp[0] = DPScore{Score: 0, PrevIdx: -1}

	for i := 1; i <= n; i++ {
		dp[i] = DPScore{Score: -1e9, PrevIdx: -1}

		for j := 0; j < i; j++ {
			phrase := Phrase(words, j, i)
			phraseLen := i - j

			// Skip single-letter phrases (but allow them in fallback)
			if phraseLen == 1 && shouldSkipWord(words[j]) {
				continue
			}

			track, score := searcher(phrase)

			if track == nil {
				continue
			}

			// Strong length bonus — prefer longer phrases exponentially
			// This ensures "I MISS YOU" (1 phrase) beats "I" + "MISS" + "YOU" (3 words)
			lengthBonus := math.Pow(2.0, float64(phraseLen)) * 20.0

			switch mode {
			case ModeExact:
				if score < 80 {
					continue
				}
				score *= 1.2
			case ModeChaos:
				if score > 60 {
					score *= 0.9
				} else {
					score *= 1.1
				}
				lengthBonus *= 0.5
			case ModeSmart:
				// Smart mode strongly prefers longer phrases
				lengthBonus *= 1.5
			}

			totalScore := dp[j].Score + score + lengthBonus

			if totalScore > dp[i].Score {
				dp[i].Score = totalScore
				dp[i].Phrase = phrase
				dp[i].Track = track
				dp[i].PrevIdx = j
				dp[i].Unavailable = false
			}
		}

		// Handle fallback cases when no valid segmentation found
		if dp[i].PrevIdx == -1 && i > 0 {
			lastValid := -1
			for k := i - 1; k >= 0; k-- {
				if dp[k].PrevIdx != -1 || k == 0 {
					lastValid = k
					break
				}
			}
			if lastValid >= 0 && lastValid < i {
				phrase := Phrase(words, lastValid, i)
				track, score := searcher(phrase)

				if mode == ModeExact && (track == nil || score < 80) {
					// Exact mode: mark segment as unavailable instead of falling back
					totalScore := dp[lastValid].Score - 10
					dp[i].Score = totalScore
					dp[i].Phrase = phrase
					dp[i].Track = nil
					dp[i].PrevIdx = lastValid
					dp[i].Unavailable = true
				} else if track != nil && score > 30 {
					totalScore := dp[lastValid].Score + score
					dp[i].Score = totalScore
					dp[i].Phrase = phrase
					dp[i].Track = track
					dp[i].PrevIdx = lastValid
					dp[i].Unavailable = false
				}
			}
		}
	}

	var tracks []Track
	idx := n
	seen := make(map[string]bool)

	for idx > 0 {
		if dp[idx].Unavailable {
			placeholder := Track{
				ID:            fmt.Sprintf("unavailable-%d", idx),
				Title:         "[unavailable]",
				Artist:        "",
				Album:         "",
				MatchType:     MatchPhrase,
				MatchScore:    0,
				MatchedPhrase: dp[idx].Phrase,
				Popularity:    0,
			}
			if !seen[placeholder.ID] {
				tracks = append(tracks, placeholder)
				seen[placeholder.ID] = true
			}
			idx = dp[idx].PrevIdx
			continue
		}
		if dp[idx].Track == nil {
			idx--
			continue
		}
		t := *dp[idx].Track
		if !seen[t.ID] {
			tracks = append(tracks, t)
			seen[t.ID] = true
		}
		idx = dp[idx].PrevIdx
	}

	for i, j := 0, len(tracks)-1; i < j; i, j = i+1, j-1 {
		tracks[i], tracks[j] = tracks[j], tracks[i]
	}

	return tracks
}

// ExactMatcher implements the strict phrase-first matching algorithm.
type ExactMatcher struct {
	searcher func(string) (*Track, float64)
	mode     Mode
	words    []string
	n        int
}

func NewExactMatcher(words []string, searcher func(string) (*Track, float64), mode Mode) *ExactMatcher {
	return &ExactMatcher{
		searcher: searcher,
		mode:     mode,
		words:    words,
		n:        len(words),
	}
}

func (em *ExactMatcher) FindBestSegmentation() []Track {
	if em.n == 0 {
		return nil
	}

	var tracks []Track
	seen := make(map[string]bool)
	covered := make(map[int]bool)

	// Try longest phrases first
	for length := em.n; length >= 1; length-- {
		for start := 0; start <= em.n-length; start++ {
			end := start + length

			alreadyCovered := false
			for i := start; i < end; i++ {
				if covered[i] {
					alreadyCovered = true
					break
				}
			}
			if alreadyCovered {
				continue
			}

			phrase := strings.Join(em.words[start:end], " ")

			if length == 1 && shouldSkipWord(em.words[start]) {
				continue
			}

			track, score := em.searcher(phrase)

			if track == nil {
				continue
			}

			if length == 1 && score < 80 {
				continue
			}

			if length > 1 && score < 60 {
				continue
			}

			if !seen[track.ID] {
				tracks = append(tracks, *track)
				seen[track.ID] = true
				for i := start; i < end; i++ {
					covered[i] = true
				}
			}
		}
	}

	// Fill uncovered words
	for i := 0; i < em.n; i++ {
		if covered[i] {
			continue
		}
		word := em.words[i]
		if shouldSkipWord(word) {
			covered[i] = true
			continue
		}
		track, score := em.searcher(word)
		if track != nil && score >= 80 && !seen[track.ID] {
			tracks = append(tracks, *track)
			seen[track.ID] = true
			covered[i] = true
		}
	}

	if len(tracks) == 0 {
		for i := 0; i < em.n; i++ {
			if covered[i] {
				continue
			}
			track, _ := em.searcher(em.words[i])
			if track != nil && !seen[track.ID] {
				tracks = append(tracks, *track)
				seen[track.ID] = true
				covered[i] = true
			}
		}
	}

	return tracks
}

func FindOptimalExact(words []string, searcher func(string) (*Track, float64), mode Mode) []Track {
	em := NewExactMatcher(words, searcher, mode)
	return em.FindBestSegmentation()
}
