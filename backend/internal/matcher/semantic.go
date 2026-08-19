package matcher

import (
	"strings"
)

// SemanticConcepts maps common emotional/semantic concepts to related keywords.
// When phrase/partial matching fails (score < 40), we check if any word in the
// phrase matches a concept, then search Spotify for related keywords.
var SemanticConcepts = map[string][]string{
	"SAD":         {"Lonely", "Tears", "Rain", "Cry", "Blue", "Melancholy", "Grief", "Sorrow", "Heartache", "Pain", "Misery", "Despair", "Hurt", "Broken", "Empty"},
	"HAPPY":       {"Joy", "Smile", "Sunshine", "Dance", "Celebration", "Cheerful", "Bliss", "Delight", "Euphoria", "Glad", "Laugh", "Fun", "Party", "Bright", "WonderFUL"},
	"ANGRY":       {"Rage", "Fury", "Mad", "Hate", "Fire", "Storm", "Revolt", "Wrath", "Furious", "Resentment", "Bitterness", "Hostility", "Conflict", "Explode", "Scream"},
	"LOVE":        {"Heart", "Kiss", "Romance", "Passion", "Darling", "Sweetheart", "Affection", "Devotion", "Adore", "Cherish", "Embrace", "Forever", "Together", "Beloved", "Desire"},
	"LONELY":      {"Alone", "Isolation", "Solitude", "Abandoned", "Forsaken", "Distant", "Remote", "Excluded", "Rejected", "Forgotten", "Unwanted", "Disconnected", "Apart", "Single", "One"},
	"TIRED":       {"Exhausted", "Weary", "Fatigue", "Sleepy", "Burned Out", "Drained", "Worn Out", "Drowsy", "Sapped", "Beat", "Spent", "Run Down", "Lethargic", "Sluggish", "Heavy"},
	"LOST":        {"Confused", "Wandering", "Adrift", "Uncertain", "Searching", "Found", "Way", "Direction", "Path", "Journey", "Unknown", "Hidden", "Mystery", "Fog", "Blind"},
	"CONFUSED":    {"Lost", "Puzzled", "Baffled", "Perplexed", "Uncertain", "Mystified", "Bewildered", "Disoriented", "Unclear", "Complex", "Chaos", "Jumble", "Muddle", "Doubt", "Question"},
	"SCARED":      {"Fear", "Terror", "Horror", "Panic", "Dread", "Fright", "Anxiety", "Nightmare", "Phobia", "Alarm", "Shock", "Threat", "Danger", "Worry", "Tremble"},
	"EXCITED":     {"Thrill", "Adventure", "Energy", "Hype", "Anticipation", "Eager", "Enthusiasm", "Exhilaration", "Electric", "Buzz", "Spark", "Rush", "Vibrate", "Alive", "Pumped"},
	"GRATEFUL":    {"Thankful", "Blessed", "Appreciate", "Grateful", "Recognition", "Acknowledgment", "Obligation", "Indebted", "Praise", "Honor", "Respect", "Tribute", "Gift", "Reward", "Grace"},
	"JEALOUS":     {"Envy", "Covet", "Possessive", "Resentful", "Green", "Suspicious", "Rival", "Competition", "Comparison", "Insecurity", "Greed", "Want", "Need", "Mine", "Theft"},
	"PROUD":       {"Achievement", "Victory", "Triumph", "Accomplishment", "Honor", "Dignity", "Confidence", "Self-Esteem", "Glory", "Success", "Win", "Champion", "Brave", "Strong", "Noble"},
	"HEARTBROKEN": {"Broken", "Shattered", "Destroyed", "Ruined", "Crushed", "Devastated", "Agony", "Betrayal", "Deception", "Abandonment", "Loss", "Grief", "Mourning", "Pain", "Scar"},
	"HOMESICK":    {"Home", "Family", "Belonging", "Nostalgia", "Missing", "Return", "Roots", "Origin", "Familiar", "Comfort", "Safe", "Warmth", "Memory", "Childhood", "Garden"},
	"STRESSED":    {"Pressure", "Tension", "Anxiety", "Overwhelmed", "Struggle", "Burden", "Weight", "Stress", "Worry", "Panic", "Deadlines", "Demands", "Strain", "Crisis", "Breakdown"},
	"HOPEFUL":     {"Hope", "Dream", "Wish", "Faith", "Believe", "Optimism", "Promise", "Future", "Light", "Dawn", "Sunrise", "New Beginning", "Possibility", "Potential", "Aspire"},
	"BORED":       {"Dull", "Monotonous", "Tedious", "Boring", "Uninteresting", "Stale", "Repetitive", "Mundane", "Routine", "Flat", "Lifeless", "Dry", "Empty", "Nothing", "Wait"},
	"NOSTALGIC":   {"Memory", "Remember", "Yesterday", "Past", "Old", "Vintage", "Retro", "Classic", "Throwback", "Childhood", "Gone", "Time", "History", "Heritage", "Legacy"},
	"BRAVE":       {"Courage", "Fearless", "Hero", "Strong", "Warrior", "Bold", "Valiant", "Daring", "Defiant", "Resilient", "Unafraid", "Steadfast", "Grit", "Determined", "Fight"},
	"ANXIOUS":     {"Nervous", "Uneasy", "Apprehensive", "Tense", "Restless", "Agitated", "Disturbed", "Troubled", "Unsettled", "Fidgety", "Jittery", "Edgy", "Worried", "Panicked", "Stressed"},
	"PEACEFUL":    {"Calm", "Serene", "Tranquil", "Quiet", "Still", "Harmony", "Gentle", "Soft", "Rest", "Relax", "Zen", "Balance", "Nature", "Ocean", "Breeze"},
	"REGRET":      {"Sorry", "Apology", "Guilt", "Remorse", "Mistake", "Fault", "Wrong", "Sorrow", "Repent", "Penance", "Shame", "Blame", "Fail", "Miss", "Wish"},
	"LONGING":     {"Yearning", "Craving", "Desire", "Want", "Need", "Hunger", "Thirst", "Pine", "Ache", "Wish", "Hope", "Dream", "Miss", "Wait", "Reach"},
	"FREE":        {"Liberty", "Independence", "Release", "Escape", "Unbound", "Wild", "Open", "Sky", "Fly", "Soar", "Unlimited", "Unchain", "Break", "Clear", "Vast"},
	"BETRAYED":    {"Lies", "Deception", "Treachery", "Backstab", "Cheat", "False", "Two-Face", "Untrust", "Disloyal", "Traitor", "Fraud", "Fake", "Trick", "Trap", "Shatter"},
	"REJECTED":    {"Denied", "Refused", "Excluded", "Outcast", "Unwanted", "Ignored", "Dismissed", "Ostracized", "Banished", "Eliminated", "Declined", "Spurned", "Abandoned", "Alone", "Nothing"},
	"GRATEFUL2":   {"Thanks", "Blessed", "Fortunate", "Lucky", "Privilege", "Wealth", "Abundance", "Plenty", "Gift", "Present", "Joy", "Happiness", "Smile", "Love", "Kind"},
	"CURIOUS":     {"Wonder", "Question", "Explore", "Discover", "Investigate", "Seek", "Learn", "Know", "Understand", "Research", "Study", "Probe", "Inquire", "Examine", "Search"},
	"FORGIVEN":    {"Mercy", "Pardon", "Absolve", "Free", "Release", "Clean", "Fresh", "New", "Start", "Heal", "Mend", "Repair", "Restore", "Renew", "Reconcile"},
}

// FindSemanticKeywords checks if any word in the phrase matches a semantic concept
// and returns related keywords for Spotify search.
func FindSemanticKeywords(phrase string) []string {
	words := strings.Fields(strings.ToUpper(phrase))
	keywordSet := make(map[string]bool)

	for _, word := range words {
		// Direct concept match
		if keywords, ok := SemanticConcepts[word]; ok {
			for _, kw := range keywords {
				keywordSet[kw] = true
			}
		}
		// Check if word contains a concept or vice versa
		for concept, keywords := range SemanticConcepts {
			if strings.Contains(word, concept) || strings.Contains(concept, word) {
				if len(word) > 2 { // Avoid matching very short substrings
					for _, kw := range keywords {
						keywordSet[kw] = true
					}
				}
			}
		}
	}

	if len(keywordSet) == 0 {
		return nil
	}

	result := make([]string, 0, len(keywordSet))
	for kw := range keywordSet {
		result = append(result, kw)
	}
	return result
}