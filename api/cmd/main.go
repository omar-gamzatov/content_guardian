package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

type ModerationRequest struct {
	TenantID      string `json:"tenant_id"`
	RequestID     string `json:"request_id"`
	ResponseMode  string `json:"response_mode"` // "sync"|"async"
	CallbackURL   string `json:"callback_url"`
	PolicyVersion string `json:"policy_version"`
	Content       struct {
		Type     string `json:"type"`
		Text     string `json:"text"`
		LangHint string `json:"lang_hint"`
	} `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

type CategoryScore struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Source string  `json:"source"` // "rule"|"model"
}

type ModelServiceResponse struct {
	Categories []CategoryScore `json:"categories"`
	Explain    map[string]any  `json:"explain"`
}

type Verdict struct {
	Action     string          `json:"action"` // "allow"|"soft_block"|"block"|"escalate"
	Severity   string          `json:"severity,omitempty"`
	Categories []CategoryScore `json:"categories"`
	Explain    map[string]any  `json:"explain"`
}

type ModerationResponse struct {
	RequestID string  `json:"request_id"`
	Verdict   Verdict `json:"verdict"`
	SLA       struct {
		LatencyMS int64  `json:"latency_ms"`
		Mode      string `json:"mode"`
	} `json:"sla"`
}

var rdb *redis.Client

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"), // "localhost:6379"
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	})
	defer rdb.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, middleware.Timeout(5*time.Second))

	r.Post("/v1/moderations", handleModeration)

	log.Println("Listening on :8080")
	http.ListenAndServe("0.0.0.0:8080", r)
}

func handleModeration(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	ctx := req.Context()

	var in ModerationRequest
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	text := normalize(in.Content.Text)
	lang := detectLang(text, in.Content.LangHint)

	cacheKey := cacheKeyFor(text, in.PolicyVersion)
	if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	// 1) Rules
	ruleScores, rulesFired := applyRules(text, lang, in.TenantID, in.PolicyVersion)

	// 2) ML (Detoxify/Presidio) — вызываем только если нужно или всегда (MVP: всегда)
	modelResponse := callModelService(ctx, text, lang)
	modelScores := modelResponse.Categories
	modelExplain := modelResponse.Explain

	// 3) Decision
	allScores := append(ruleScores, modelScores...)
	verdict := decide(allScores, loadThresholds(in.TenantID, in.PolicyVersion), modelExplain, rulesFired)

	resp := ModerationResponse{
		RequestID: in.RequestID,
		Verdict:   verdict,
	}
	resp.SLA.LatencyMS = time.Since(start).Milliseconds()
	resp.SLA.Mode = pickMode(in.ResponseMode)

	out, _ := json.Marshal(resp)
	rdb.Set(ctx, cacheKey, out, 10*time.Minute) // кэшируй коротко; настройка per-tenant

	// TODO: persist в Postgres (заявка+вердикт), опубликовать событие в NATS при async

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func cacheKeyFor(text string, policyVer string) string {
	sum := sha256.Sum256([]byte(text + "|" + policyVer))
	return "mcache:" + hex.EncodeToString(sum[:])
}

// ----- Stub implementations ниже -----

func normalize(s string) string { return s } // lowercasing, unicode NFKC, trim, collapse spaces

func detectLang(text, hint string) string {
	if hint != "" {
		return hint
	}
	// TODO: fastText or heuristic
	return "en"
}

type Thresholds map[string]struct{ Block, Soft, Allow float64 }

func loadThresholds(tenant, policyVersion string) Thresholds {
	// TODO: load from Policy service / file / DB
	return Thresholds{
		"toxicity":        {Block: 0.92, Soft: 0.75, Allow: 0.3},
		"sexual_explicit": {Block: 0.9, Soft: 0.7, Allow: 0.2},
		"identity_attack": {Block: 0.9, Soft: 0.7, Allow: 0.3},
		"violence_threat": {Block: 0.9, Soft: 0.7, Allow: 0.3},
		"profanity":       {Block: 0.95, Soft: 0.8, Allow: 0.0},
	}
}

func applyRules(text, lang, tenant, policyVer string) ([]CategoryScore, []string) {
	// TODO: parse and evaluate YAML DSL
	var scores []CategoryScore
	var fired []string
	// Dummy: если содержит "kill you"
	if containsInsensitive(text, "kill you") {
		scores = append(scores, CategoryScore{Name: "violence_threat", Score: 0.85, Source: "rule"})
		fired = append(fired, "threat_keywords_v1")
	}
	return scores, fired
}

func callModelService(ctx context.Context, text, lang string) ModelServiceResponse {
	// HTTP POST http://ml:8000/classify {text, lang}
	data := map[string]string{
		"text": text,
		"lang": lang,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(
		"http://ml:8000/classify",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("Error: %s\n", string(err.Error()))
		return ModelServiceResponse{}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", string(body))
	var response ModelServiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}
	// Parse JSON with scores
	// Stub:
	return response
}

func decide(scores []CategoryScore, th Thresholds, modelExplain map[string]any, rulesFired []string) Verdict {
	action := "allow"
	sev := "low"
	maxScore := 0.0
	for _, s := range scores {
		t := th[s.Name]
		if s.Score >= t.Block {
			action = maxAction(action, "block")
		} else if s.Score >= t.Soft {
			action = maxAction(action, "soft_block")
		}
		if s.Score > maxScore {
			maxScore = s.Score
		}
	}
	uncertainty := 1 - maxScore
	if uncertainty > 0.15 && action != "block" {
		action = "escalate"
	}
	if action == "block" || action == "escalate" {
		sev = "high"
	} else if action == "soft_block" {
		sev = "medium"
	}

	return Verdict{
		Action:     action,
		Severity:   sev,
		Categories: aggregate(scores),
		Explain: map[string]any{
			"rules_fired":    rulesFired,
			"model":          modelExplain,
			"uncertainty":    uncertainty,
			"policy_version": "v1.0",
		},
	}
}
func maxAction(a, b string) string {
	order := map[string]int{"allow": 0, "soft_block": 1, "block": 2, "escalate": 3}
	if order[b] > order[a] {
		return b
	}
	return a
}
func aggregate(scores []CategoryScore) []CategoryScore {
	m := map[string]CategoryScore{}
	for _, s := range scores {
		if cur, ok := m[s.Name]; !ok || s.Score > cur.Score {
			m[s.Name] = s
		}
	}
	out := make([]CategoryScore, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func containsInsensitive(text, sub string) bool {
	// TODO: implement properly
	return false
}

func pickMode(m string) string {
	if m == "async" {
		return "async"
	}
	return "sync"
}
