package handler

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// ──────────────────────────────────────────────────────────────────────────────
// MODELS
// ──────────────────────────────────────────────────────────────────────────────

type AttachmentStyle string

const (
	Secure       AttachmentStyle = "secure"
	Anxious      AttachmentStyle = "anxious"
	Avoidant     AttachmentStyle = "avoidant"
	Disorganized AttachmentStyle = "disorganized"
)

type StyleInfo struct {
	Style     AttachmentStyle
	Label     string
	Tagline   string
	Color     string
	BgColor   string
	Summary   string
	Strengths []string
	Growth    string
	Note      string
}

var Styles = map[AttachmentStyle]StyleInfo{
	Secure: {
		Style: Secure, Label: "Secure", Tagline: "The safe harbour",
		Color: "#4A8C7A", BgColor: "#F0F5F2",
		Summary:   "You move through relationships with quiet confidence. You're comfortable with closeness and equally at ease with independence. Conflict doesn't mean collapse, and vulnerability is a doorway, not a weakness.",
		Strengths: []string{"Easy to trust and be trusted", "Emotionally available without losing yourself", "Resilient through conflict and repair"},
		Growth:    "Your stability can sometimes make it harder to understand partners in earlier stages of healing. Explicit curiosity about their inner world goes a long way.",
		Note:      "Secure attachment is a strength earned through experience — and it can always be deepened.",
	},
	Avoidant: {
		Style: Avoidant, Label: "Avoidant", Tagline: "The independent soul",
		Color: "#7A6E9E", BgColor: "#F2F0F5",
		Summary:   "You've learned to be extraordinarily self-sufficient and you protect that fiercely. Closeness can feel like a slow loss of self, so you keep your inner world private. This isn't coldness — it's armour that once served you well.",
		Strengths: []string{"Deeply self-aware and independent", "Clear and respected personal boundaries", "Calm and composed in a crisis"},
		Growth:    "Letting someone in doesn't have to mean losing yourself. The intimacy you protect yourself from might also be what you're quietly longing for.",
		Note:      "Avoidant patterns often form from learning early that needing others was unsafe. That wisdom made sense once.",
	},
	Anxious: {
		Style: Anxious, Label: "Anxious", Tagline: "The devoted heart",
		Color: "#B06060", BgColor: "#F5F0F0",
		Summary:   "You love deeply and feel deeply — sometimes achingly so. Your antennae for emotional shifts are finely tuned. The worry underneath isn't weakness; it's how much you care, paired with the old fear that it might not be enough.",
		Strengths: []string{"Deeply empathic and attuned to others", "Wholeheartedly present in relationships", "Courageously and genuinely vulnerable"},
		Growth:    "Learning to soothe yourself — rather than seeking external reassurance — is the work. You are not too much. But your wholeness can't depend entirely on another.",
		Note:      "Anxious attachment often develops when love felt inconsistent early on. The hypervigilance was adaptive.",
	},
	Disorganized: {
		Style: Disorganized, Label: "Disorganized", Tagline: "The complex navigator",
		Color: "#B59050", BgColor: "#F5F2EE",
		Summary:   "You carry contradictions that can feel exhausting: the pull toward closeness and the urge to flee. Relationships feel like walking a tightrope. This isn't brokenness — it's the signature of someone who needed safety in places that were unpredictable.",
		Strengths: []string{"Profound self-knowledge through lived experience", "Deep capacity for empathy toward others in pain", "Nuanced, non-binary emotional intelligence"},
		Growth:    "Consistency — in yourself and in chosen relationships — is the medicine. Therapy and relationships with secure people can genuinely rewire old patterns.",
		Note:      "Disorganized attachment is the most complex style, and the most responsive to intentional healing.",
	},
}

type Option struct {
	Label string
	Value string
}

type Question struct {
	ID       int
	Category string
	Text     string
	Options  []Option
}

type ShuffledQuestion struct {
	Question
	ShuffledOptions []Option
}

type ProgressMetric struct {
	Week          int
	SecurityScore float64
	MoodAvg       float64
	CheckIns      int
}

var Questions = []Question{
	{1, "closeness", "When someone you care about grows very close to you, your instinct is to…", []Option{
		{"Lean in — closeness feels safe and wonderful", "secure"},
		{"Feel warmth, but keep some parts of yourself private", "avoidant"},
		{"Worry they'll eventually pull away", "anxious"},
		{"Feel uncertain — sometimes warm, sometimes overwhelmed", "disorganized"},
	}},
	{2, "conflict", "After a disagreement with someone close to you, you typically…", []Option{
		{"Talk it through calmly and move on without lingering doubt", "secure"},
		{"Need space to process alone before you can reconnect", "avoidant"},
		{"Replay the conversation, wondering if they're still upset", "anxious"},
		{"Feel conflicted — wanting to fix it but also wanting to flee", "disorganized"},
	}},
	{3, "independence", "When a partner or close friend becomes more independent, you feel…", []Option{
		{"Happy for them — space is healthy and natural", "secure"},
		{"Relieved — you value your own independence equally", "avoidant"},
		{"A little anxious — you wonder if they need you less", "anxious"},
		{"Mixed — partly relieved, partly scared of being abandoned", "disorganized"},
	}},
	{4, "vulnerability", "Sharing a deep fear or insecurity with someone close feels…", []Option{
		{"Connecting — vulnerability genuinely strengthens bonds", "secure"},
		{"Uncomfortable — you prefer to handle your inner world privately", "avoidant"},
		{"Risky — what if they judge or use it against you?", "anxious"},
		{"Both necessary and terrifying at the same time", "disorganized"},
	}},
	{5, "reassurance", "How often do you seek reassurance from people you're close to?", []Option{
		{"Occasionally, and I feel fine asking for it", "secure"},
		{"Rarely — I don't like feeling like I need it", "avoidant"},
		{"Often — a little more certainty always helps", "anxious"},
		{"Sometimes desperately, sometimes I push it away entirely", "disorganized"},
	}},
	{6, "self-worth", "Deep down, your sense of being loveable feels…", []Option{
		{"Pretty solid — I know my worth regardless of others", "secure"},
		{"Fine, but tied to my accomplishments and self-sufficiency", "avoidant"},
		{"A bit fragile — I often need confirmation from others", "anxious"},
		{"Inconsistent — some days clear, other days completely lost", "disorganized"},
	}},
	{7, "ideal relationship", "When you imagine an ideal close relationship, it would feel…", []Option{
		{"Warm, honest, and comfortable with both closeness and space", "secure"},
		{"Respectful of boundaries, low-pressure, intellectually engaging", "avoidant"},
		{"Deep, devoted, and emotionally available — always present", "anxious"},
		{"Safe — somehow both intense and stable at the same time", "disorganized"},
	}},
}

// ──────────────────────────────────────────────────────────────────────────────
// AI COACH
// ──────────────────────────────────────────────────────────────────────────────

func translateMessage(original string, style AttachmentStyle) string {
	original = strings.TrimSpace(original)
	if original == "" {
		return ""
	}
	lower := strings.ToLower(original)
	switch style {
	case Anxious:
		if containsAny(lower, "you never", "you always", "you don't care") {
			return "Try: \"When this happens, I feel disconnected and I really miss feeling close to you. Can we talk about it?\"\n\nNote: Your original message may land as an accusation. Reframing as a need invites connection rather than defence."
		}
		if containsAny(lower, "fine", "whatever", "forget it") {
			return "Try: \"I'm not fine — I'm feeling hurt and I'm shutting down. I need a few minutes and then I'd like to reconnect.\"\n\nNote: Underneath dismissive language is usually a genuine need. Naming it gives your partner something to respond to."
		}
		if containsAny(lower, "do you love", "do you still", "are we okay") {
			return "Try: \"I'm feeling uncertain about us right now and I could really use some reassurance. Are you available to connect?\"\n\nNote: Making the need explicit removes guesswork and is easier for a partner to answer."
		}
		return "Your message carries genuine feeling. Consider adding: \"What I need right now is ___.\" Naming the need directly reduces anxiety for both of you.\n\nOriginal: \"" + original + "\""
	case Avoidant:
		if containsAny(lower, "i need space", "i need time", "leave me alone") {
			return "Try: \"I care about you and I'm feeling overwhelmed right now. I need a couple of hours to myself — and then I'd like to come back to this.\"\n\nNote: Adding a return signal prevents your partner from experiencing your withdrawal as abandonment."
		}
		if containsAny(lower, "i'm fine", "i'm okay", "it doesn't matter") {
			return "Try: \"I'm finding it hard to put this into words right now, but I'll try: I feel ___ and I need ___.\"\n\nNote: Even a small window of access goes a long way."
		}
		return "Consider adding one small emotional disclosure: \"I feel ___ when this happens.\" It doesn't have to be big — just one word. It helps your partner feel seen.\n\nOriginal: \"" + original + "\""
	case Disorganized:
		return "Your message may contain mixed signals (approach + avoidance). Try grounding it in one clear request: \"Right now I need ___ from you.\" If that feels impossible, name the conflict itself: \"Part of me wants to be close and part of me wants to run — I'm working through it.\"\n\nOriginal: \"" + original + "\""
	default:
		return "Your message reads clearly. To make it even more connecting, consider adding what you'd like to happen next: \"What would help me most is ___.\"\n\nOriginal: \"" + original + "\""
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func dailyPrompt(style AttachmentStyle, day int) string {
	prompts := map[AttachmentStyle][]string{
		Secure:       {"Today: Notice one moment when you felt genuinely connected. What made it feel safe?", "Today: Practice naming one emotion out loud — even to yourself.", "Today: Ask someone a question you're genuinely curious about."},
		Anxious:      {"Today: When you feel the urge to seek reassurance, pause for 90 seconds first. What do you actually need?", "Today: Write one thing you genuinely appreciate about yourself — no external validation needed.", "Today: Notice one moment of connection you might have dismissed as 'not enough.'"},
		Avoidant:     {"Today: Share one small, low-stakes feeling with someone you trust.", "Today: When you feel the urge to withdraw, notice it without acting on it for 2 minutes.", "Today: Name one thing another person did that you appreciated — even silently."},
		Disorganized: {"Today: When you feel conflicted about closeness, write both sides down. Both are valid.", "Today: Notice one moment of safety — no matter how small.", "Today: Practice one grounding breath before responding to a trigger."},
	}
	list := prompts[style]
	if len(list) == 0 {
		return "Today: Notice one moment of genuine connection."
	}
	return list[day%len(list)]
}

func securityScore(scores map[string]int, total int) int {
	if total == 0 {
		return 0
	}
	s := scores["secure"]
	partial := scores["avoidant"]/3 + scores["anxious"]/3
	raw := float64(s+partial) / float64(total)
	if raw > 1 {
		raw = 1
	}
	return int(raw * 100)
}

func dynamicAnalysis(a, b AttachmentStyle) string {
	key := string(a) + "-" + string(b)
	if v, ok := dynamicsMap[key]; ok {
		return v
	}
	key = string(b) + "-" + string(a)
	if v, ok := dynamicsMap[key]; ok {
		return v
	}
	return "This pairing holds unique dynamics. Each partner brings their own history — awareness and curiosity are the most powerful tools available to you both."
}

var dynamicsMap = map[string]string{
	"anxious-avoidant":     "The Anxious-Avoidant Trap: One of the most common and painful relational patterns. The anxious partner's bids for closeness trigger the avoidant partner's withdrawal — which amplifies the anxious partner's fear and pursuit. Both are responding rationally to their own nervous systems, but the cycle feeds itself. The path forward requires the anxious partner to develop self-soothing skills and the avoidant partner to practice tolerating closeness incrementally.",
	"secure-anxious":       "The Secure-Anxious Dynamic: A secure partner can be profoundly stabilising for an anxious one — but only if the secure partner doesn't become a substitute for the anxious partner's own self-regulation. The risk: co-dependency dressed as security. The gift: the anxious partner experiences consistent, available love — perhaps for the first time — which gradually rewires their baseline expectation of relationships.",
	"secure-avoidant":      "The Secure-Avoidant Dynamic: A secure partner gives the avoidant space without abandoning them, which can be deeply disorienting (in the best way) for the avoidant partner. The risk: the secure partner eventually feeling chronically under-nurtured. The gift: the avoidant partner slowly learns that closeness doesn't erase selfhood.",
	"secure-secure":        "The Secure-Secure Dynamic: The gold standard — not perfect, but resilient. These partners have enough internal stability to repair quickly, give each other genuine space, and grow individually while growing together.",
	"anxious-anxious":      "The Anxious-Anxious Dynamic: Two highly attuned, deeply feeling people — which can create extraordinary intimacy. The risk: mutual amplification of fear and enmeshment. The gift: unparalleled emotional understanding between partners who truly see each other.",
	"avoidant-avoidant":    "The Avoidant-Avoidant Dynamic: Two independent people who deeply respect space — which can feel effortless until emotional depth is required. The risk: a relationship that is companionate but emotionally shallow. The gift: genuine mutual respect for autonomy.",
	"disorganized-secure":  "The Disorganized-Secure Dynamic: Potentially the most healing pairing for a disorganized partner. A secure person's consistency — staying present through the push-pull — can be transformative. The risk: the secure partner must maintain strong boundaries and not treat the relationship as a healing project.",
	"disorganized-anxious": "The Disorganized-Anxious Dynamic: Two partners with activated nervous systems — intense closeness, but also intense friction. Both need significant self-regulation work for this pairing to thrive. The gift: neither partner pathologises the other's fear.",
	"disorganized-avoidant": "The Disorganized-Avoidant Dynamic: The disorganized partner's push-pull dynamic meets the avoidant partner's withdrawal — creating cycles of pursuit, confusion, and distance. Clear communication agreements are essential.",
}

func triggerAlertResponse(style AttachmentStyle, score int) string {
	base := "High stress detected. "
	switch style {
	case Anxious:
		return base + "Your nervous system is activated. Before reaching out: try 4-7-8 breathing (inhale 4 counts, hold 7, exhale 8). Then identify: what do I actually need right now?"
	case Avoidant:
		return base + "You may feel an urge to disconnect. Try setting a specific return time (\"I'll be back in 30 minutes\") rather than going fully dark. That one signal prevents a lot of relational damage."
	case Disorganized:
		return base + "Mixed impulses are normal right now. Grounding first: name 5 things you can see. Then: you don't have to act on any impulse in the next 10 minutes."
	default:
		return base + "Take a moment. What do you need right now — connection, space, or just to breathe? You have access to all three."
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// TEMPLATE ENGINE
// ──────────────────────────────────────────────────────────────────────────────

var (
	tmpl  *template.Template
	store = sessions.NewCookieStore([]byte(getSecret()))
)

func getSecret() string {
	if s := os.Getenv("SESSION_SECRET"); s != "" {
		return s
	}
	return "attachsecure-dev-secret-32chars!!"
}

func findTemplatesDir() string {
	// Search candidates from most to least likely
	candidates := []string{
		"templates",
		"../templates",
		"../../templates",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "layout.html")); err == nil {
			return c
		}
	}
	return "templates" // fallback
}

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"pct": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a * 100 / b
		},
		"intVal":   func(f float64) int { return int(f) },
		"styleFor": func(s string) StyleInfo { return Styles[AttachmentStyle(s)] },
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
		},
	}

	dir := findTemplatesDir()
	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseGlob(filepath.Join(dir, "*.html"))
	if err != nil {
		log.Printf("Template parse error: %v", err)
	}
}

func render(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	if tmpl == nil {
		http.Error(w, "Templates not loaded — check server logs", 500)
		return
	}
	sess, _ := store.Get(r, "as-session")
	data["CurrentPath"] = r.URL.Path
	data["HasResult"] = sess.Values["quiz_done"]
	data["SessionStyle"] = sess.Values["style"]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template execute error (%s): %v", name, err)
		http.Error(w, "Render error", 500)
	}
}

func sessInt(sess *sessions.Session, key string) int {
	v, _ := sess.Values[key].(int)
	return v
}

// ──────────────────────────────────────────────────────────────────────────────
// ROUTER
// ──────────────────────────────────────────────────────────────────────────────

var router *mux.Router

func init() {
	router = mux.NewRouter()
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)
	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/quiz", quizHandler).Methods("GET")
	router.HandleFunc("/quiz/submit", quizSubmitHandler).Methods("POST")
	router.HandleFunc("/result", resultHandler).Methods("GET")
	router.HandleFunc("/dashboard", dashboardHandler).Methods("GET")
	router.HandleFunc("/coach", coachHandler).Methods("GET")
	router.HandleFunc("/coach/translate", coachTranslateHandler).Methods("POST")
	router.HandleFunc("/checkin", checkinHandler).Methods("GET", "POST")
	router.HandleFunc("/about", aboutHandler).Methods("GET")
	router.HandleFunc("/prd", prdHandler).Methods("GET")
}

// Handler is the Vercel serverless entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

// ──────────────────────────────────────────────────────────────────────────────
// HANDLERS
// ──────────────────────────────────────────────────────────────────────────────

func homeHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "home.html", map[string]interface{}{
		"Title": "AttachSecure — Understand Your Attachment Style",
	})
}

func quizHandler(w http.ResponseWriter, r *http.Request) {
	qs := make([]ShuffledQuestion, len(Questions))
	for i, q := range Questions {
		opts := make([]Option, len(q.Options))
		copy(opts, q.Options)
		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		qs[i] = ShuffledQuestion{q, opts}
	}
	render(w, r, "quiz.html", map[string]interface{}{
		"Title":     "The Attachment Assessment — AttachSecure",
		"Questions": qs,
		"Total":     len(Questions),
	})
}

func quizSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", 400)
		return
	}
	scores := map[string]int{"secure": 0, "anxious": 0, "avoidant": 0, "disorganized": 0}
	total := 0
	for i := range Questions {
		val := r.FormValue("q" + strconv.Itoa(i+1))
		if val != "" {
			scores[val]++
			total++
		}
	}
	dominant, max := "secure", -1
	for style, count := range scores {
		if count > max {
			max = count
			dominant = style
		}
	}
	ss := securityScore(scores, total)
	sess, _ := store.Get(r, "as-session")
	sess.Values["style"] = dominant
	sess.Values["security_score"] = ss
	sess.Values["scores_secure"] = scores["secure"]
	sess.Values["scores_anxious"] = scores["anxious"]
	sess.Values["scores_avoidant"] = scores["avoidant"]
	sess.Values["scores_disorganized"] = scores["disorganized"]
	sess.Values["total"] = total
	sess.Values["user_id"] = uuid.New().String()
	sess.Values["quiz_done"] = true
	sess.Save(r, w)
	http.Redirect(w, r, "/result", http.StatusSeeOther)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		http.Redirect(w, r, "/quiz", http.StatusSeeOther)
		return
	}
	secScore := sessInt(sess, "security_score")
	total := sessInt(sess, "total")
	if total == 0 {
		total = len(Questions)
	}
	styleInfo := Styles[AttachmentStyle(style)]
	type bar struct {
		Label string
		Count int
		Color string
		Pct   int
	}
	breakdown := []bar{
		{"Secure", sessInt(sess, "scores_secure"), "#4A8C7A", sessInt(sess, "scores_secure") * 100 / total},
		{"Anxious", sessInt(sess, "scores_anxious"), "#B06060", sessInt(sess, "scores_anxious") * 100 / total},
		{"Avoidant", sessInt(sess, "scores_avoidant"), "#7A6E9E", sessInt(sess, "scores_avoidant") * 100 / total},
		{"Disorganized", sessInt(sess, "scores_disorganized"), "#B59050", sessInt(sess, "scores_disorganized") * 100 / total},
	}
	render(w, r, "result.html", map[string]interface{}{
		"Title":          "Your Attachment Style — AttachSecure",
		"Style":          style,
		"StyleInfo":      styleInfo,
		"SecurityScore":  secScore,
		"ScoreBreakdown": breakdown,
		"TodayPrompt":    dailyPrompt(AttachmentStyle(style), time.Now().YearDay()),
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = "anxious"
	}
	secScore := sessInt(sess, "security_score")
	if secScore == 0 {
		secScore = 42
	}
	weekly := []ProgressMetric{
		{1, 35, 5.2, 3},
		{2, 38, 5.8, 5},
		{3, 41, 6.1, 6},
		{4, float64(secScore), 6.4, 7},
	}
	render(w, r, "dashboard.html", map[string]interface{}{
		"Title":       "Your Dashboard — AttachSecure",
		"Style":       style,
		"StyleInfo":   Styles[AttachmentStyle(style)],
		"SecScore":    secScore,
		"WeeklyData":  weekly,
		"TodayPrompt": dailyPrompt(AttachmentStyle(style), time.Now().YearDay()),
		"QuizDone":    sess.Values["quiz_done"],
	})
}

func coachHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = "anxious"
	}
	partnerStyle := r.URL.Query().Get("partner")
	var analysis string
	if partnerStyle != "" {
		analysis = dynamicAnalysis(AttachmentStyle(style), AttachmentStyle(partnerStyle))
	}
	render(w, r, "coach.html", map[string]interface{}{
		"Title":           "AI Communication Coach — AttachSecure",
		"Style":           style,
		"StyleInfo":       Styles[AttachmentStyle(style)],
		"PartnerStyle":    partnerStyle,
		"DynamicAnalysis": analysis,
		"StyleOptions":    []string{"secure", "anxious", "avoidant", "disorganized"},
	})
}

func coachTranslateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = r.FormValue("style")
	}
	msg := r.FormValue("message")
	translated := translateMessage(msg, AttachmentStyle(style))
	partnerStyle := r.FormValue("partner_style")
	var analysis string
	if partnerStyle != "" {
		analysis = dynamicAnalysis(AttachmentStyle(style), AttachmentStyle(partnerStyle))
	}
	render(w, r, "coach.html", map[string]interface{}{
		"Title":           "AI Communication Coach — AttachSecure",
		"Style":           style,
		"StyleInfo":       Styles[AttachmentStyle(style)],
		"Original":        msg,
		"Translated":      translated,
		"PartnerStyle":    partnerStyle,
		"DynamicAnalysis": analysis,
		"StyleOptions":    []string{"secure", "anxious", "avoidant", "disorganized"},
	})
}

func checkinHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = "anxious"
	}
	type entry struct {
		Score    int
		Trigger  string
		Note     string
		IsStress bool
	}
	var alert string
	var logged *entry
	if r.Method == "POST" {
		r.ParseForm()
		score, _ := strconv.Atoi(r.FormValue("score"))
		trigger := r.FormValue("trigger")
		note := r.FormValue("note")
		isStress := score <= 3
		logged = &entry{score, trigger, note, isStress}
		if isStress {
			alert = triggerAlertResponse(AttachmentStyle(style), score)
		}
	}
	render(w, r, "checkin.html", map[string]interface{}{
		"Title":       "Daily Check-In — AttachSecure",
		"Style":       style,
		"StyleInfo":   Styles[AttachmentStyle(style)],
		"TodayPrompt": dailyPrompt(AttachmentStyle(style), time.Now().YearDay()),
		"Alert":       alert,
		"Entry":       logged,
	})
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "about.html", map[string]interface{}{"Title": "About — AttachSecure"})
}

func prdHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "prd.html", map[string]interface{}{"Title": "Product Requirements Document — AttachSecure"})
}
