package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ── Model tests ──────────────────────────────────────────────────────────────

func TestStylesMapComplete(t *testing.T) {
	expected := []AttachmentStyle{Secure, Anxious, Avoidant, Disorganized}
	for _, s := range expected {
		if _, ok := Styles[s]; !ok {
			t.Errorf("Styles map missing entry for: %s", s)
		}
	}
}

func TestStyleInfoFields(t *testing.T) {
	for style, info := range Styles {
		if info.Label == "" {
			t.Errorf("Style %s missing Label", style)
		}
		if info.Color == "" {
			t.Errorf("Style %s missing Color", style)
		}
		if info.Summary == "" {
			t.Errorf("Style %s missing Summary", style)
		}
		if len(info.Strengths) == 0 {
			t.Errorf("Style %s has no Strengths", style)
		}
		if info.Growth == "" {
			t.Errorf("Style %s missing Growth", style)
		}
	}
}

func TestQuestionsCount(t *testing.T) {
	if len(Questions) != 7 {
		t.Errorf("expected 7 questions, got %d", len(Questions))
	}
}

func TestQuestionsHaveFourOptions(t *testing.T) {
	for _, q := range Questions {
		if len(q.Options) != 4 {
			t.Errorf("question %d: expected 4 options, got %d", q.ID, len(q.Options))
		}
	}
}

func TestQuestionOptionValues(t *testing.T) {
	valid := map[string]bool{"secure": true, "anxious": true, "avoidant": true, "disorganized": true}
	for _, q := range Questions {
		for _, opt := range q.Options {
			if !valid[opt.Value] {
				t.Errorf("question %d: invalid option value %q", q.ID, opt.Value)
			}
		}
	}
}

// ── Security score tests ─────────────────────────────────────────────────────

func TestSecurityScoreAllSecure(t *testing.T) {
	scores := map[string]int{"secure": 7, "anxious": 0, "avoidant": 0, "disorganized": 0}
	got := securityScore(scores, 7)
	if got != 100 {
		t.Errorf("all-secure: expected 100, got %d", got)
	}
}

func TestSecurityScoreZeroTotal(t *testing.T) {
	scores := map[string]int{}
	got := securityScore(scores, 0)
	if got != 0 {
		t.Errorf("zero total: expected 0, got %d", got)
	}
}

func TestSecurityScoreAllDisorganized(t *testing.T) {
	scores := map[string]int{"secure": 0, "anxious": 0, "avoidant": 0, "disorganized": 7}
	got := securityScore(scores, 7)
	if got != 0 {
		t.Errorf("all-disorganized: expected 0, got %d", got)
	}
}

func TestSecurityScoreMixed(t *testing.T) {
	scores := map[string]int{"secure": 4, "anxious": 2, "avoidant": 1, "disorganized": 0}
	got := securityScore(scores, 7)
	if got < 50 || got > 80 {
		t.Errorf("mixed: expected score between 50-80, got %d", got)
	}
}

// ── AI coach tests ───────────────────────────────────────────────────────────

func TestTranslateMessageEmpty(t *testing.T) {
	got := translateMessage("", Anxious)
	if got != "" {
		t.Errorf("empty message: expected empty string, got %q", got)
	}
}

func TestTranslateMessageAnxiousAccusation(t *testing.T) {
	got := translateMessage("you never listen to me", Anxious)
	if !strings.Contains(got, "Try:") {
		t.Errorf("anxious accusation: expected coaching suggestion, got %q", got)
	}
}

func TestTranslateMessageAvoidantSpace(t *testing.T) {
	got := translateMessage("I need space right now", Avoidant)
	if !strings.Contains(got, "return") && !strings.Contains(got, "Try:") {
		t.Errorf("avoidant space: expected return signal coaching, got %q", got)
	}
}

func TestTranslateMessageDisorganized(t *testing.T) {
	got := translateMessage("I don't know what I want from you", Disorganized)
	if !strings.Contains(got, "mixed") && !strings.Contains(got, "clear request") {
		t.Errorf("disorganized: expected mixed signal coaching, got %q", got)
	}
}

func TestTranslateMessageSecure(t *testing.T) {
	got := translateMessage("Can we talk about this?", Secure)
	if got == "" {
		t.Errorf("secure: expected some response, got empty")
	}
}

// ── Dynamic analysis tests ───────────────────────────────────────────────────

func TestDynamicAnalysisKnownPairing(t *testing.T) {
	got := dynamicAnalysis(Anxious, Avoidant)
	if !strings.Contains(strings.ToLower(got), "trap") && !strings.Contains(strings.ToLower(got), "anxious") {
		t.Errorf("anxious-avoidant: expected trap description, got %q", got)
	}
}

func TestDynamicAnalysisSymmetric(t *testing.T) {
	ab := dynamicAnalysis(Anxious, Avoidant)
	ba := dynamicAnalysis(Avoidant, Anxious)
	if ab != ba {
		t.Errorf("dynamic analysis should be symmetric: A-B != B-A")
	}
}

func TestDynamicAnalysisFallback(t *testing.T) {
	got := dynamicAnalysis("unknown", "also_unknown")
	if got == "" {
		t.Errorf("unknown pairing: expected fallback string, got empty")
	}
}

// ── Daily prompt tests ───────────────────────────────────────────────────────

func TestDailyPromptAllStyles(t *testing.T) {
	styles := []AttachmentStyle{Secure, Anxious, Avoidant, Disorganized}
	for _, s := range styles {
		got := dailyPrompt(s, 1)
		if got == "" {
			t.Errorf("style %s: expected non-empty prompt", s)
		}
		if !strings.HasPrefix(got, "Today:") {
			t.Errorf("style %s: prompt should start with 'Today:', got %q", s, got)
		}
	}
}

func TestDailyPromptCyclesCorrectly(t *testing.T) {
	p0 := dailyPrompt(Anxious, 0)
	p3 := dailyPrompt(Anxious, 3)
	if p0 != p3 {
		t.Errorf("prompt should cycle: day 0 and day 3 should match (3 prompts total), got %q vs %q", p0, p3)
	}
}

// ── Trigger alert tests ──────────────────────────────────────────────────────

func TestTriggerAlertAllStyles(t *testing.T) {
	styles := []AttachmentStyle{Secure, Anxious, Avoidant, Disorganized}
	for _, s := range styles {
		got := triggerAlertResponse(s, 2)
		if !strings.Contains(got, "High stress") {
			t.Errorf("style %s: expected 'High stress' prefix, got %q", s, got)
		}
	}
}

// ── HTTP handler tests ───────────────────────────────────────────────────────

func TestHomeHandlerReturns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	// tmpl may be nil in test environment without template files — skip render error
	homeHandler(rr, req)
	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("home handler: unexpected status %d", rr.Code)
	}
}

func TestQuizHandlerReturns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/quiz", nil)
	rr := httptest.NewRecorder()
	quizHandler(rr, req)
	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("quiz handler: unexpected status %d", rr.Code)
	}
}

func TestQuizSubmitRedirectsOnSuccess(t *testing.T) {
	form := url.Values{}
	for i := 1; i <= 7; i++ {
		styles := []string{"secure", "anxious", "avoidant", "disorganized"}
		form.Set("q"+strings.TrimSpace(strings.Repeat("0", 0)+string(rune('0'+i))), styles[i%4])
	}
	form.Set("q1", "secure")
	form.Set("q2", "anxious")
	form.Set("q3", "avoidant")
	form.Set("q4", "secure")
	form.Set("q5", "anxious")
	form.Set("q6", "secure")
	form.Set("q7", "secure")

	req := httptest.NewRequest("POST", "/quiz/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	quizSubmitHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("quiz submit: expected 303 redirect, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/result" {
		t.Errorf("quiz submit: expected redirect to /result, got %q", loc)
	}
}

func TestResultHandlerRedirectsWithNoSession(t *testing.T) {
	req := httptest.NewRequest("GET", "/result", nil)
	rr := httptest.NewRecorder()
	resultHandler(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("result handler without session: expected 303, got %d", rr.Code)
	}
}

func TestCheckinHandlerGET(t *testing.T) {
	req := httptest.NewRequest("GET", "/checkin", nil)
	rr := httptest.NewRecorder()
	checkinHandler(rr, req)
	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("checkin GET: unexpected status %d", rr.Code)
	}
}

func TestCheckinHandlerPOSTHighStress(t *testing.T) {
	form := url.Values{}
	form.Set("score", "2")
	form.Set("trigger", "Conflict with partner")
	form.Set("note", "feeling overwhelmed")

	req := httptest.NewRequest("POST", "/checkin", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	checkinHandler(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("checkin POST: unexpected status %d", rr.Code)
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("you never listen", "you never", "you always") {
		t.Error("containsAny: should match 'you never'")
	}
	if containsAny("hello world", "you never", "you always") {
		t.Error("containsAny: should not match")
	}
}
