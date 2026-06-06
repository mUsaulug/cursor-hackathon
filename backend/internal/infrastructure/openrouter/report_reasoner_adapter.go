package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

// DefaultModel is a cheap, fast model good enough for short report prose.
const DefaultModel = "openai/gpt-4o-mini"

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

// Reasoner calls OpenRouter to rewrite the maintenance summary in natural
// Turkish. It NEVER makes KVKK/priority decisions: those come from the
// deterministic LocalReasoner base, and only the summary prose is replaced.
// On any error it falls back to the deterministic report (graceful degradation).
type Reasoner struct {
	apiKey   string
	model    string
	client   *http.Client
	fallback appvision.ReasonerPort
}

// NewReasoner builds the OpenRouter reasoner over a deterministic fallback.
func NewReasoner(apiKey, model string, fallback appvision.ReasonerPort) *Reasoner {
	if model == "" {
		model = DefaultModel
	}
	if fallback == nil {
		fallback = NewLocalReasoner()
	}
	return &Reasoner{
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 12 * time.Second},
		fallback: fallback,
	}
}

// GenerateMaintenanceReport produces the deterministic base report, then asks
// the LLM to improve only the summary sentence. Determinism is preserved for
// risk_level, recommended_action, and kvkk_note.
func (r *Reasoner) GenerateMaintenanceReport(ctx context.Context, result domain.AnalysisResult) (*appvision.MaintenanceReport, error) {
	base, err := r.fallback.GenerateMaintenanceReport(ctx, result)
	if err != nil || base == nil {
		return r.fallback.GenerateMaintenanceReport(ctx, result)
	}

	prose, perr := r.summarize(ctx, result, base)
	if perr != nil || prose == "" {
		return base, nil // graceful: keep deterministic summary
	}
	base.Summary = prose
	return base, nil
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (r *Reasoner) summarize(ctx context.Context, result domain.AnalysisResult, base *appvision.MaintenanceReport) (string, error) {
	var types []string
	for _, d := range result.Detections {
		types = append(types, fmt.Sprintf("%s (%s, %s)", d.NormalizedObjectType, d.Priority, d.ReviewStatus))
	}
	user := fmt.Sprintf(
		"Tespitler: %v. Risk: %s. Bu verilere dayanarak tek cumlelik, Turkce, resmi bir belediye bakim ozeti yaz. Karar verme, sadece ozetle.",
		types, base.RiskLevel,
	)

	body, _ := json.Marshal(chatRequest{
		Model: r.model,
		Messages: []chatMessage{
			{Role: "system", Content: "Sen bir belediye bakim raporu yazarisi. Yalnizca verilen tespitleri ozetlersin; oncelik veya KVKK kararlarini asla degistirmezsin."},
			{Role: "user", Content: user},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return "", fmt.Errorf("openrouter: status %d: %s", resp.StatusCode, b)
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("openrouter: empty choices")
	}
	return cr.Choices[0].Message.Content, nil
}
