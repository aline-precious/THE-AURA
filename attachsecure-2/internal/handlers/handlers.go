package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"attachsecure/internal/ai"
	"attachsecure/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("attachsecure-secret-key-change-in-prod"))
var templates *template.Template

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
}

func RegisterRoutes(r *mux.Router) {
	// Parse all templates once
	var err error
	templates, err = template.New("").Funcs(template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"sub":      func(a, b int) int { return a - b },
		"mul":      func(a, b int) int { return a * b },
		"pct":      func(a, b int) int { if b == 0 { return 0 }; return a * 100 / b },
		"styleInfo": func(s string) models.StyleInfo { return models.Styles[models.AttachmentStyle(s)] },
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"nl2br":    func(s string) template.HTML { return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>")) },
	}).ParseGlob("templates/**/*.html")
	if err != nil {
		// Try flat glob for flexibility
		templates, err = template.New("").Funcs(template.FuncMap{
			"add":      func(a, b int) int { return a + b },
			"sub":      func(a, b int) int { return a - b },
			"mul":      func(a, b int) int { return a * b },
			"pct":      func(a, b int) int { if b == 0 { return 0 }; return a * 100 / b },
			"styleInfo": func(s string) models.StyleInfo { return models.Styles[models.AttachmentStyle(s)] },
			"safeHTML": func(s string) template.HTML { return template.HTML(s) },
			"nl2br":    func(s string) template.HTML { return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>")) },
		}).ParseGlob("templates/*.html")
		if err != nil {
			log.Printf("Warning: could not parse templates: %v", err)
		}
	}

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Pages
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/quiz", QuizHandler).Methods("GET")
	r.HandleFunc("/quiz/submit", QuizSubmitHandler).Methods("POST")
	r.HandleFunc("/result", ResultHandler).Methods("GET")
	r.HandleFunc("/dashboard", DashboardHandler).Methods("GET")
	r.HandleFunc("/coach", CoachHandler).Methods("GET")
	r.HandleFunc("/coach/translate", CoachTranslateHandler).Methods("POST")
	r.HandleFunc("/checkin", CheckInHandler).Methods("GET", "POST")
	r.HandleFunc("/about", AboutHandler).Methods("GET")
	r.HandleFunc("/prd", PRDHandler).Methods("GET")

	// API
	r.HandleFunc("/api/mood", MoodAPIHandler).Methods("POST")
	r.HandleFunc("/api/dynamic", DynamicAPIHandler).Methods("GET")
}

// --- Page Handlers ---

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "home.html", map[string]interface{}{
		"Title": "AttachSecure — Understand Your Attachment Style",
	})
}

func QuizHandler(w http.ResponseWriter, r *http.Request) {
	// Shuffle options for each question
	type ShuffledQuestion struct {
		models.Question
		ShuffledOptions []models.Option
	}
	qs := make([]ShuffledQuestion, len(models.Questions))
	for i, q := range models.Questions {
		opts := make([]models.Option, len(q.Options))
		copy(opts, q.Options)
		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		qs[i] = ShuffledQuestion{q, opts}
	}
	renderTemplate(w, r, "quiz.html", map[string]interface{}{
		"Title":     "The Attachment Quiz — AttachSecure",
		"Questions": qs,
		"Total":     len(models.Questions),
	})
}

func QuizSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	scores := map[string]int{"secure": 0, "anxious": 0, "avoidant": 0, "disorganized": 0}
	total := 0
	for i := range models.Questions {
		key := "q" + strconv.Itoa(i+1)
		val := r.FormValue(key)
		if val != "" {
			scores[val]++
			total++
		}
	}

	// Determine dominant style
	dominant := "secure"
	max := -1
	for style, count := range scores {
		if count > max {
			max = count
			dominant = style
		}
	}

	secScore := ai.SecurityScore(scores, total)

	// Store in session
	sess, _ := store.Get(r, "as-session")
	sess.Values["style"] = dominant
	sess.Values["security_score"] = secScore
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

func ResultHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		http.Redirect(w, r, "/quiz", http.StatusSeeOther)
		return
	}

	secScore, _ := sess.Values["security_score"].(int)
	total, _ := sess.Values["total"].(int)
	styleInfo := models.Styles[models.AttachmentStyle(style)]
	if total == 0 {
		total = len(models.Questions)
	}

	scoreBreakdown := []struct {
		Style string
		Label string
		Count int
		Color string
		Pct   int
	}{
		{"secure", "Secure", getInt(sess, "scores_secure"), "#4A8C7A", getInt(sess, "scores_secure") * 100 / total},
		{"anxious", "Anxious", getInt(sess, "scores_anxious"), "#B06060", getInt(sess, "scores_anxious") * 100 / total},
		{"avoidant", "Avoidant", getInt(sess, "scores_avoidant"), "#7A6E9E", getInt(sess, "scores_avoidant") * 100 / total},
		{"disorganized", "Disorganized", getInt(sess, "scores_disorganized"), "#B59050", getInt(sess, "scores_disorganized") * 100 / total},
	}

	prompt := ai.DailyPrompt(models.AttachmentStyle(style), time.Now().YearDay())

	renderTemplate(w, r, "result.html", map[string]interface{}{
		"Title":          "Your Attachment Style — AttachSecure",
		"Style":          style,
		"StyleInfo":      styleInfo,
		"SecurityScore":  secScore,
		"ScoreBreakdown": scoreBreakdown,
		"TodayPrompt":    prompt,
	})
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = "anxious" // demo default
	}
	styleInfo := models.Styles[models.AttachmentStyle(style)]
	secScore, _ := sess.Values["security_score"].(int)
	if secScore == 0 {
		secScore = 42
	}

	// Demo weekly progress data
	weeklyData := []models.ProgressMetric{
		{Week: 1, SecurityScore: 35, MoodAvg: 5.2, CheckIns: 3},
		{Week: 2, SecurityScore: 38, MoodAvg: 5.8, CheckIns: 5},
		{Week: 3, SecurityScore: 41, MoodAvg: 6.1, CheckIns: 6},
		{Week: 4, SecurityScore: secScore, MoodAvg: 6.4, CheckIns: 7},
	}

	prompt := ai.DailyPrompt(models.AttachmentStyle(style), time.Now().YearDay())

	renderTemplate(w, r, "dashboard.html", map[string]interface{}{
		"Title":       "Your Dashboard — AttachSecure",
		"Style":       style,
		"StyleInfo":   styleInfo,
		"SecScore":    secScore,
		"WeeklyData":  weeklyData,
		"TodayPrompt": prompt,
		"QuizDone":    sess.Values["quiz_done"],
	})
}

func CoachHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = "anxious"
	}

	// Demo interaction analysis
	partnerStyle := r.URL.Query().Get("partner")
	var dynamicAnalysis string
	if partnerStyle != "" {
		dynamicAnalysis = ai.DynamicAnalysis(models.AttachmentStyle(style), models.AttachmentStyle(partnerStyle))
	}

	renderTemplate(w, r, "coach.html", map[string]interface{}{
		"Title":           "AI Communication Coach — AttachSecure",
		"Style":           style,
		"StyleInfo":       models.Styles[models.AttachmentStyle(style)],
		"PartnerStyle":    partnerStyle,
		"DynamicAnalysis": dynamicAnalysis,
		"StyleOptions":    []string{"secure", "anxious", "avoidant", "disorganized"},
	})
}

func CoachTranslateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", 400)
		return
	}
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = r.FormValue("style")
	}
	msg := r.FormValue("message")
	translated := ai.TranslateMessage(msg, models.AttachmentStyle(style))

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="translation-result">` + strings.ReplaceAll(template.HTMLEscapeString(translated), "\n", "<br>") + `</div>`))
		return
	}

	renderTemplate(w, r, "coach.html", map[string]interface{}{
		"Title":       "AI Communication Coach — AttachSecure",
		"Style":       style,
		"StyleInfo":   models.Styles[models.AttachmentStyle(style)],
		"Original":    msg,
		"Translated":  translated,
		"StyleOptions": []string{"secure", "anxious", "avoidant", "disorganized"},
	})
}

func CheckInHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)
	if style == "" {
		style = "anxious"
	}

	var alert string
	var entry *models.MoodEntry

	if r.Method == "POST" {
		r.ParseForm()
		scoreStr := r.FormValue("score")
		score, _ := strconv.Atoi(scoreStr)
		trigger := r.FormValue("trigger")
		note := r.FormValue("note")

		entry = &models.MoodEntry{
			ID:          uuid.New().String(),
			UserID:      "demo",
			Score:       score,
			Trigger:     trigger,
			IsHighStress: score <= 3,
			Note:        note,
			CreatedAt:   time.Now(),
		}

		if entry.IsHighStress {
			alert = ai.TriggerAlertResponse(models.AttachmentStyle(style), score)
		}
	}

	prompt := ai.DailyPrompt(models.AttachmentStyle(style), time.Now().YearDay())

	renderTemplate(w, r, "checkin.html", map[string]interface{}{
		"Title":       "Daily Check-In — AttachSecure",
		"Style":       style,
		"StyleInfo":   models.Styles[models.AttachmentStyle(style)],
		"TodayPrompt": prompt,
		"Alert":       alert,
		"Entry":       entry,
	})
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "about.html", map[string]interface{}{
		"Title": "About AttachSecure",
	})
}

func PRDHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "prd.html", map[string]interface{}{
		"Title": "Product Requirements Document — AttachSecure",
	})
}

// --- API Handlers ---

func MoodAPIHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Score   int    `json:"score"`
		Trigger string `json:"trigger"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	sess, _ := store.Get(r, "as-session")
	style, _ := sess.Values["style"].(string)

	resp := map[string]interface{}{
		"id":        uuid.New().String(),
		"high_stress": req.Score <= 3,
		"alert":    "",
	}
	if req.Score <= 3 {
		resp["alert"] = ai.TriggerAlertResponse(models.AttachmentStyle(style), req.Score)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func DynamicAPIHandler(w http.ResponseWriter, r *http.Request) {
	a := r.URL.Query().Get("a")
	b := r.URL.Query().Get("b")
	analysis := ai.DynamicAnalysis(models.AttachmentStyle(a), models.AttachmentStyle(b))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"analysis": analysis})
}

// --- Helpers ---

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	if templates == nil {
		http.Error(w, "Templates not loaded", 500)
		return
	}
	sess, _ := store.Get(r, "as-session")
	data["CurrentPath"] = r.URL.Path
	data["HasQuizResult"] = sess.Values["quiz_done"]
	data["SessionStyle"] = sess.Values["style"]

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template error (%s): %v", name, err)
		http.Error(w, "Template error", 500)
	}
}

func getInt(sess *sessions.Session, key string) int {
	v, _ := sess.Values[key].(int)
	return v
}
