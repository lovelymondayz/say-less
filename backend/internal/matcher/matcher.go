package matcher

import (
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

	replacements := map[string]string{
		"DON'T": "DO NOT", "WON'T": "WILL NOT", "CAN'T": "CANNOT",
		"ISN'T": "IS NOT", "AREN'T": "ARE NOT", "WASN'T": "WAS NOT",
		"WEREN'T": "WERE NOT", "HAVEN'T": "HAVE NOT", "HASN'T": "HAS NOT",
		"HADN'T": "HAD NOT", "DOESN'T": "DOES NOT", "DIDN'T": "DID NOT",
		"I'M": "I AM", "I'VE": "I HAVE", "I'LL": "I WILL", "I'D": "I WOULD",
		"YOU'RE": "YOU ARE", "YOU'VE": "YOU HAVE", "YOU'LL": "YOU WILL",
		"HE'S": "HE IS", "SHE'S": "SHE IS", "IT'S": "IT IS",
		"WE'RE": "WE ARE", "THEY'RE": "THEY ARE",
	}

	for contraction, expanded := range replacements {
		s = strings.ReplaceAll(s, contraction, expanded)
	}

	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}

	return strings.Join(strings.Fields(b.String()), " ")
}

func SplitWords(text string) []string {
	return strings.Fields(text)
}

// FindBestSegmentation finds the fewest songs that reconstruct the message.
// It searches the full phrase first, then progressively shorter segments.
// From the pool of results, it picks songs whose titles contain the words.
func FindBestSegmentation(words []string, searcher func(string) ([]Track, float64), mode Mode) []Track {
	if len(words) == 0 {
		return nil
	}

	fullPhrase := strings.Join(words, " ")
	
	// Search the full phrase first to get a pool of results
	poolTracks, _ := searcher(fullPhrase)
	
	// If full phrase search returned results, try to find one that covers everything
	if len(poolTracks) > 0 {
		for _, t := range poolTracks {
			if coversFullMessage(t.Title, fullPhrase) {
				return []Track{t}
			}
		}
	}

	// Progressive segmentation: try longest phrases first
	// Search each unique phrase length once, collect all results
	allResults := make(map[string][]Track)
	
	for length := len(words); length >= 1; length-- {
		for start := 0; start <= len(words)-length; start++ {
			end := start + length
			phrase := strings.Join(words[start:end], " ")
			
			if _, searched := allResults[phrase]; !searched {
				tracks, _ := searcher(phrase)
				allResults[phrase] = tracks
			}
		}
	}

	// DP to find best segmentation
	n := len(words)
	
	type state struct {
		score  float64
		tracks []Track
		used   map[string]bool
	}
	
	dp := make([]state, n+1)
	dp[0] = state{score: 0, tracks: nil, used: make(map[string]bool)}
	
	for i := 1; i <= n; i++ {
		dp[i] = state{score: -1, used: make(map[string]bool)}
		
		for j := 0; j < i; j++ {
			if dp[j].score < 0 {
				continue
			}
			
			phrase := strings.Join(words[j:i], " ")
			tracks := allResults[phrase]
			
			for _, t := range tracks {
				if dp[j].used[t.ID] {
					continue
				}
				
				// Score: phrase length bonus + match quality
				phraseLen := i - j
				lengthBonus := math.Pow(2.0, float64(phraseLen)) * 20.0
				
				// Check if track title actually contains the phrase
				trackUpper := strings.ToUpper(t.Title)
				if !strings.Contains(trackUpper, phrase) && !strings.Contains(phrase, trackUpper) {
					// Check word overlap
					phraseWords := strings.Fields(phrase)
					trackWords := strings.Fields(trackUpper)
					matched := 0
					for _, pw := range phraseWords {
						for _, tw := range trackWords {
							if pw == tw {
								matched++
								break
							}
						}
					}
					if matched < len(phraseWords)/2+1 {
						continue
					}
				}
				
				newScore := dp[j].score + lengthBonus + t.MatchScore
				
				if newScore > dp[i].score {
					newTracks := make([]Track, len(dp[j].tracks))
					copy(newTracks, dp[j].tracks)
					newTracks = append(newTracks, t)
					
					newUsed := make(map[string]bool)
					for k, v := range dp[j].used {
						newUsed[k] = v
					}
					newUsed[t.ID] = true
					
					dp[i] = state{
						score:  newScore,
						tracks: newTracks,
						used:   newUsed,
					}
				}
			}
		}
	}
	
	if dp[n].score >= 0 {
		return dp[n].tracks
	}
	
	// Fallback: return whatever we have
	return dp[n].tracks
}

func coversFullMessage(title, message string) bool {
	titleUpper := strings.ToUpper(title)
	msgUpper := strings.ToUpper(message)
	
	if strings.Contains(titleUpper, msgUpper) {
		return true
	}
	
	msgWords := strings.Fields(msgUpper)
	matched := 0
	for _, mw := range msgWords {
		if strings.Contains(titleUpper, mw) {
			matched++
		}
	}
	return matched >= len(msgWords)*3/4
}

func FindOptimalSegmentation(words []string, searcher func(string) ([]Track, float64), mode Mode) []Track {
	return FindBestSegmentation(words, searcher, mode)
}

func FindOptimalExact(words []string, searcher func(string) ([]Track, float64), mode Mode) []Track {
	fullPhrase := strings.Join(words, " ")
	tracks, _ := searcher(fullPhrase)
	
	for _, t := range tracks {
		if strings.Contains(strings.ToUpper(t.Title), fullPhrase) {
			return []Track{t}
		}
	}
	
	return FindBestSegmentation(words, searcher, mode)
}
