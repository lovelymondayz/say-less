package matcher

import (
	"strings"
)

// ExactMatcher implements the strict phrase-first matching algorithm.
// It prioritizes finding the fewest tracks that exactly match phrases.
// No semantic fallback, no popularity boost.
type ExactMatcher struct {
	searcher func(string) (*Track, float64)
	mode     Mode
	words    []string
	n        int
}

// NewExactMatcher creates a new strict phrase-first matcher.
func NewExactMatcher(words []string, searcher func(string) (*Track, float64), mode Mode) *ExactMatcher {
	return &ExactMatcher{
		searcher: searcher,
		mode:     mode,
		words:    words,
		n:        len(words),
	}
}

// FindBestSegmentation finds the optimal phrase-first segmentation.
// It prefers longer phrases with exact matches over shorter ones.
func (em *ExactMatcher) FindBestSegmentation() []Track {
	if em.n == 0 {
		return nil
	}

	// Try longest phrases first, then work down
	var tracks []Track
	seen := make(map[string]bool)
	covered := make(map[int]bool) // Track which word positions are covered

	// First pass: try to find exact matches for the longest possible phrases
	for length := em.n; length >= 1; length-- {
		for start := 0; start <= em.n-length; start++ {
			end := start + length

			// Check if any word in this range is already covered
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

			// Skip single-letter words in phrases (like "I", "A")
			if length == 1 && len(phrase) <= 1 {
				continue
			}

			track, score := em.searcher(phrase)

			// Only accept good matches
			if track == nil {
				continue
			}

			// For single words, only accept exact matches
			if length == 1 && score < 80 {
				continue
			}

			// For phrases, prefer strong matches
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

	// Fill uncovered words with single-word matches or unavailable
	for i := 0; i < em.n; i++ {
		if covered[i] {
			continue
		}
		word := em.words[i]
		if len(word) <= 1 {
			// Mark as covered but don't add a track
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

	// If still nothing, use any available tracks
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

// FindOptimalExact is the strict phrase-first algorithm.
// Unlike FindOptimalSegmentation, it greedily selects the longest matching phrases first.
func FindOptimalExact(words []string, searcher func(string) (*Track, float64), mode Mode) []Track {
	em := NewExactMatcher(words, searcher, mode)
	return em.FindBestSegmentation()
}
