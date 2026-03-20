package ai

import (
	"fmt"
	"strings"
	"attachsecure/internal/models"
)

// TranslateMessage reframes a message based on the sender's attachment style
// to reduce defensiveness in communication.
func TranslateMessage(original string, style models.AttachmentStyle) string {
	original = strings.TrimSpace(original)
	if original == "" {
		return ""
	}

	switch style {
	case models.Anxious:
		return translateAnxious(original)
	case models.Avoidant:
		return translateAvoidant(original)
	case models.Disorganized:
		return translateDisorganized(original)
	default:
		return fmt.Sprintf("✓ Your message reads clearly and securely: \"%s\"", original)
	}
}

func translateAnxious(msg string) string {
	lower := strings.ToLower(msg)

	if contains(lower, "you never", "you always", "you don't care") {
		return fmt.Sprintf(
			"Try: \"When this happens, I feel disconnected and I really miss feeling close to you. Can we talk about it?\""+
				"\n\n(Your original message: \"%s\" — may land as an accusation. Reframing as a need invites connection rather than defence.)", msg)
	}
	if contains(lower, "fine", "whatever", "forget it") {
		return fmt.Sprintf(
			"Try: \"I'm not fine — I'm feeling hurt and I'm shutting down. I need a few minutes and then I'd like to reconnect.\""+
				"\n\n(Underneath \"%s\" is usually a genuine need. Naming it gives your partner something to respond to.)", msg)
	}
	if contains(lower, "do you love", "do you still", "are we okay") {
		return fmt.Sprintf(
			"Try: \"I'm feeling uncertain about us right now and I could really use some reassurance. Are you available to connect?\""+
				"\n\n(Your message expresses a real need. Making it explicit removes the guesswork and is easier for a partner to answer.)", msg)
	}

	return fmt.Sprintf(
		"Your message carries genuine feeling. Consider adding: \"What I need right now is ___.\" "+
			"Naming the need directly reduces anxiety for both of you.\n\nOriginal: \"%s\"", msg)
}

func translateAvoidant(msg string) string {
	lower := strings.ToLower(msg)

	if contains(lower, "i need space", "i need time", "leave me alone") {
		return fmt.Sprintf(
			"Try: \"I care about you and I'm feeling overwhelmed right now. I need a couple of hours to myself — and then I'd like to come back to this.\""+
				"\n\n(Adding a return signal to \"%s\" prevents your partner from experiencing your withdrawal as abandonment.)", msg)
	}
	if contains(lower, "i'm fine", "i'm okay", "it doesn't matter") {
		return fmt.Sprintf(
			"Try: \"I'm finding it hard to put this into words right now, but I'll try: I'm feeling ___ and I need ___.\""+
				"\n\n(\"%s\" often signals shutdown. Even a small window of access goes a long way.)", msg)
	}

	return fmt.Sprintf(
		"Consider adding one small emotional disclosure: \"I feel ___ when this happens.\" "+
			"It doesn't have to be big — just one word. It helps your partner feel seen.\n\nOriginal: \"%s\"", msg)
}

func translateDisorganized(msg string) string {
	return fmt.Sprintf(
		"Your message may contain mixed signals (approach + avoidance). Try grounding it in one clear request: "+
			"\"Right now I need ___ from you.\" If that feels impossible, name the conflict itself: "+
			"\"Part of me wants to be close and part of me wants to run — I'm working through it.\"\n\nOriginal: \"%s\"", msg)
}

func contains(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

// DailyPrompt returns a micro-interaction prompt based on attachment style
func DailyPrompt(style models.AttachmentStyle, day int) string {
	prompts := map[models.AttachmentStyle][]string{
		models.Secure: {
			"Today: Notice one moment when you felt genuinely connected to someone. What made it feel safe?",
			"Today: Practice naming one emotion out loud — even to yourself.",
			"Today: Ask someone a question you're genuinely curious about.",
		},
		models.Anxious: {
			"Today: When you feel the urge to seek reassurance, pause for 90 seconds first. What do you actually need?",
			"Today: Write one thing you genuinely appreciate about yourself — no external validation needed.",
			"Today: Notice one moment of connection you might have dismissed as 'not enough.'",
		},
		models.Avoidant: {
			"Today: Share one small, low-stakes feeling with someone you trust.",
			"Today: When you feel the urge to withdraw, notice it without acting on it for 2 minutes.",
			"Today: Name one thing another person did today that you appreciated — even silently.",
		},
		models.Disorganized: {
			"Today: When you feel conflicted about closeness, write both sides down. Both are valid.",
			"Today: Notice one moment of safety — no matter how small.",
			"Today: Practice one grounding breath before responding to a trigger.",
		},
	}
	list := prompts[style]
	if len(list) == 0 {
		return "Today: Notice one moment of genuine connection."
	}
	return list[day%len(list)]
}

// SecurityScore calculates a simple score 0-100 based on answer distribution
func SecurityScore(scores map[string]int, total int) int {
	if total == 0 {
		return 0
	}
	secureCount := scores["secure"]
	// Partial credit: avoidant and anxious partially toward security
	partialCount := scores["avoidant"]/3 + scores["anxious"]/3
	raw := float64(secureCount+partialCount) / float64(total)
	if raw > 1 {
		raw = 1
	}
	return int(raw * 100)
}

// DynamicAnalysis describes the interaction pattern between two styles
func DynamicAnalysis(styleA, styleB models.AttachmentStyle) string {
	key := string(styleA) + "-" + string(styleB)
	if v, ok := dynamics[key]; ok {
		return v
	}
	key = string(styleB) + "-" + string(styleA)
	if v, ok := dynamics[key]; ok {
		return v
	}
	return "This pairing holds unique dynamics. Each partner brings their own history — awareness and curiosity are the most powerful tools."
}

var dynamics = map[string]string{
	"anxious-avoidant": "The Anxious-Avoidant Trap: This is one of the most common and painful relational patterns. The anxious partner's bids for closeness trigger the avoidant partner's withdrawal — which in turn amplifies the anxious partner's fear and pursuit. Both are responding rationally to their own nervous systems, but the cycle feeds itself. The path forward requires the anxious partner to develop self-soothing and the avoidant partner to practice tolerating closeness incrementally.",
	"secure-anxious":   "The Secure-Anxious Dynamic: A secure partner can be profoundly stabilising for an anxious one — but only if the secure partner doesn't become a substitute for the anxious partner's own self-regulation. The risk: co-dependency dressed as security. The gift: the anxious partner experiences consistent, available love — perhaps for the first time — which gradually rewires their baseline expectation of relationships.",
	"secure-avoidant":  "The Secure-Avoidant Dynamic: A secure partner gives the avoidant space without abandoning them, which can be deeply disorienting (in the best way) for the avoidant partner. The risk: the secure partner eventually feeling chronically under-nurtured. The gift: the avoidant partner slowly learns that closeness doesn't erase selfhood.",
	"secure-secure":    "The Secure-Secure Dynamic: The gold standard — not perfect, but resilient. These partners have enough internal stability to repair quickly, give each other genuine space, and grow individually while growing together.",
	"anxious-anxious":  "The Anxious-Anxious Dynamic: Two highly attuned, deeply feeling people — which can create extraordinary intimacy. The risk: mutual amplification of fear, co-regulation without individual regulation, and enmeshment. The gift: unparalleled emotional understanding between partners.",
	"avoidant-avoidant": "The Avoidant-Avoidant Dynamic: Two independent people who respect space — which can feel effortless until emotional depth is required. The risk: a relationship that is companionate but emotionally shallow. The gift: genuine mutual respect for autonomy.",
	"disorganized-secure": "The Disorganized-Secure Dynamic: This may be the most healing pairing for a disorganized partner. A secure person's consistency — staying present through the push-pull — can be transformative. The risk: the secure partner must have strong boundaries and not treat the relationship as a healing project.",
}

// TriggerAlertResponse returns UI guidance text for a high-stress event
func TriggerAlertResponse(style models.AttachmentStyle, score int) string {
	base := fmt.Sprintf("High stress detected (score: %d/10). ", score)
	switch style {
	case models.Anxious:
		return base + "Your nervous system is activated. Before reaching out to your partner: try 4-7-8 breathing (inhale 4 counts, hold 7, exhale 8). Then identify: what do I actually need right now?"
	case models.Avoidant:
		return base + "You may feel an urge to disconnect. Try: set a specific time (\"I'll be back in 30 minutes\") rather than going fully dark. That one signal prevents a lot of relational damage."
	case models.Disorganized:
		return base + "Mixed impulses are normal right now. Grounding first: name 5 things you can see. Then: you don't have to act on any impulse in the next 10 minutes."
	default:
		return base + "Take a moment. What do you need right now — connection, space, or just to breathe? You have access to all three."
	}
}
