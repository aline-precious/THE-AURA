package models

import "time"

type AttachmentStyle string

const (
	Secure       AttachmentStyle = "secure"
	Anxious      AttachmentStyle = "anxious"
	Avoidant     AttachmentStyle = "avoidant"
	Disorganized AttachmentStyle = "disorganized"
)

type UserRole string

const (
	RoleSeeker   UserRole = "seeker"
	RoleCouple   UserRole = "couple"
	RoleTherapist UserRole = "therapist"
)

type User struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Email          string          `json:"email"`
	Role           UserRole        `json:"role"`
	AttachmentStyle AttachmentStyle `json:"attachment_style"`
	SecurityScore  int             `json:"security_score"` // 0-100
	JoinedAt       time.Time       `json:"joined_at"`
	PartnerID      string          `json:"partner_id,omitempty"`
}

type QuizAnswer struct {
	QuestionID int    `json:"question_id"`
	Value      string `json:"value"` // secure | anxious | avoidant | disorganized
}

type QuizResult struct {
	UserID          string          `json:"user_id"`
	Style           AttachmentStyle `json:"style"`
	Scores          map[string]int  `json:"scores"`
	CompletedAt     time.Time       `json:"completed_at"`
	SecurityPercent int             `json:"security_percent"`
}

type MoodEntry struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Score       int       `json:"score"` // 1-10
	Trigger     string    `json:"trigger,omitempty"`
	IsHighStress bool     `json:"is_high_stress"`
	Note        string    `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CoachMessage struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Original    string    `json:"original"`
	Translated  string    `json:"translated"`
	Style       AttachmentStyle `json:"style"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProgressMetric struct {
	Week          int     `json:"week"`
	SecurityScore float64 `json:"security_score"`
	MoodAvg       float64 `json:"mood_avg"`
	CheckIns      int     `json:"check_ins"`
}

type StyleInfo struct {
	Style    AttachmentStyle
	Label    string
	Tagline  string
	Color    string
	BgColor  string
	Summary  string
	Strengths []string
	Growth   string
	Note     string
}

var Styles = map[AttachmentStyle]StyleInfo{
	Secure: {
		Style:   Secure,
		Label:   "Secure",
		Tagline: "The safe harbour",
		Color:   "#4A8C7A",
		BgColor: "#F0F5F2",
		Summary: "You move through relationships with quiet confidence. You're comfortable with closeness and equally at ease with independence. Conflict doesn't mean collapse, and vulnerability is a doorway, not a weakness.",
		Strengths: []string{
			"Easy to trust and be trusted",
			"Emotionally available without losing yourself",
			"Resilient through conflict and repair",
		},
		Growth: "Your stability can sometimes make it harder to understand partners in earlier stages of healing. Explicit curiosity about their inner world goes a long way.",
		Note:   "Secure attachment is a strength earned through experience — and it can always be deepened.",
	},
	Avoidant: {
		Style:   Avoidant,
		Label:   "Avoidant",
		Tagline: "The independent soul",
		Color:   "#7A6E9E",
		BgColor: "#F2F0F5",
		Summary: "You've learned to be extraordinarily self-sufficient and you protect that fiercely. Closeness can feel like a slow loss of self, so you keep your inner world private. This isn't coldness — it's armour that once served you well.",
		Strengths: []string{
			"Deeply self-aware and independent",
			"Clear and respected personal boundaries",
			"Calm and composed in a crisis",
		},
		Growth: "Letting someone in doesn't have to mean losing yourself. The intimacy you protect yourself from might also be what you're quietly longing for.",
		Note:   "Avoidant patterns often form from learning early that needing others was unsafe. That wisdom made sense once.",
	},
	Anxious: {
		Style:   Anxious,
		Label:   "Anxious",
		Tagline: "The devoted heart",
		Color:   "#B06060",
		BgColor: "#F5F0F0",
		Summary: "You love deeply and feel deeply — sometimes achingly so. Your antennae for emotional shifts are finely tuned. The worry underneath isn't weakness; it's how much you care, paired with the old fear that it might not be enough.",
		Strengths: []string{
			"Deeply empathic and attuned to others",
			"Wholeheartedly present in relationships",
			"Courageously and genuinely vulnerable",
		},
		Growth: "Learning to soothe yourself — rather than seeking external reassurance — is the work. You are not too much. But your wholeness can't depend entirely on another.",
		Note:   "Anxious attachment often develops when love felt inconsistent early on. The hypervigilance was adaptive.",
	},
	Disorganized: {
		Style:   Disorganized,
		Label:   "Disorganized",
		Tagline: "The complex navigator",
		Color:   "#B59050",
		BgColor: "#F5F2EE",
		Summary: "You carry contradictions that can feel exhausting: the pull toward closeness and the urge to flee. Relationships feel like walking a tightrope. This isn't brokenness — it's the signature of someone who needed safety in places that were unpredictable.",
		Strengths: []string{
			"Profound self-knowledge through lived experience",
			"Deep capacity for empathy toward others in pain",
			"Nuanced, non-binary emotional intelligence",
		},
		Growth: "Consistency — in yourself and in chosen relationships — is the medicine. Therapy and relationships with secure people can genuinely rewire old patterns.",
		Note:   "Disorganized attachment is the most complex style, and the most responsive to intentional healing.",
	},
}

type Question struct {
	ID       int
	Category string
	Text     string
	Options  []Option
}

type Option struct {
	Label string
	Value string
}

var Questions = []Question{
	{
		ID: 1, Category: "closeness",
		Text: "When someone you care about grows very close to you, your instinct is to…",
		Options: []Option{
			{"Lean in — closeness feels safe and wonderful", "secure"},
			{"Feel warmth, but keep some parts of yourself private", "avoidant"},
			{"Worry they'll eventually pull away", "anxious"},
			{"Feel uncertain — sometimes warm, sometimes overwhelmed", "disorganized"},
		},
	},
	{
		ID: 2, Category: "conflict",
		Text: "After a disagreement with someone close to you, you typically…",
		Options: []Option{
			{"Talk it through calmly and move on without lingering doubt", "secure"},
			{"Need space to process alone before you can reconnect", "avoidant"},
			{"Replay the conversation, wondering if they're still upset", "anxious"},
			{"Feel conflicted — wanting to fix it but also wanting to flee", "disorganized"},
		},
	},
	{
		ID: 3, Category: "independence",
		Text: "When a partner or close friend becomes more independent, you feel…",
		Options: []Option{
			{"Happy for them — space is healthy and natural", "secure"},
			{"Relieved — you value your own independence equally", "avoidant"},
			{"A little anxious — you wonder if they need you less", "anxious"},
			{"Mixed — partly relieved, partly scared of being abandoned", "disorganized"},
		},
	},
	{
		ID: 4, Category: "vulnerability",
		Text: "Sharing a deep fear or insecurity with someone close feels…",
		Options: []Option{
			{"Connecting — vulnerability genuinely strengthens bonds", "secure"},
			{"Uncomfortable — you prefer to handle your inner world privately", "avoidant"},
			{"Risky — what if they judge or use it against you?", "anxious"},
			{"Both necessary and terrifying at the same time", "disorganized"},
		},
	},
	{
		ID: 5, Category: "reassurance",
		Text: "How often do you seek reassurance from people you're close to?",
		Options: []Option{
			{"Occasionally, and I feel fine asking for it", "secure"},
			{"Rarely — I don't like feeling like I need it", "avoidant"},
			{"Often — a little more certainty always helps", "anxious"},
			{"Sometimes desperately, sometimes I push it away entirely", "disorganized"},
		},
	},
	{
		ID: 6, Category: "self-worth",
		Text: "Deep down, your sense of being loveable feels…",
		Options: []Option{
			{"Pretty solid — I know my worth regardless of others", "secure"},
			{"Fine, but tied to my accomplishments and self-sufficiency", "avoidant"},
			{"A bit fragile — I often need confirmation from others", "anxious"},
			{"Inconsistent — some days clear, other days completely lost", "disorganized"},
		},
	},
	{
		ID: 7, Category: "ideal relationship",
		Text: "When you imagine an ideal close relationship, it would feel…",
		Options: []Option{
			{"Warm, honest, and comfortable with both closeness and space", "secure"},
			{"Respectful of boundaries, low-pressure, intellectually engaging", "avoidant"},
			{"Deep, devoted, and emotionally available — always present", "anxious"},
			{"Safe — somehow both intense and stable at the same time", "disorganized"},
		},
	},
}
