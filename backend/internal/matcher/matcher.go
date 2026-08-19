package matcher

import (
	"fmt"
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

type DPScore struct {
	Score      float64
	Phrase     string
	Track      *Track
	PrevIdx    int
	Unavailable bool
}

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
			track, score := searcher(phrase)

			if track == nil {
				continue
			}

			lengthBonus := float64(i-j) * 5.0
			
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
				lengthBonus *= 1.0
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
					totalScore := dp[lastValid].Score - 10 // penalty
					dp[i].Score = totalScore
					dp[i].Phrase = phrase
					dp[i].Track = nil // Will be handled in backtracking
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
			// Insert placeholder for unavailable segment
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
