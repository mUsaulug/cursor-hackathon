// Package vision is the application layer of the CivicLens Core vision bounded
// context. It orchestrates the deterministic decision chain
// (normalize -> privacy guard -> confidence -> priority -> review) and depends
// only on the domain layer and outbound port interfaces. Adapters live in
// internal/infrastructure and implement these ports (hexagonal architecture).
package vision

import (
	"context"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// InferenceInput is the adapter-agnostic input for a single inference call.
type InferenceInput struct {
	Image      []byte
	SourceType string
	SourceRef  string
}

// InferencePort is implemented by each perception adapter (DETR live,
// precomputed, road-damage). It returns raw, un-normalized detections; the
// CivicLens Core pipeline does all normalization and policy.
type InferencePort interface {
	Detect(ctx context.Context, in InferenceInput) ([]domain.RawDetection, error)
	ModelID() string
	ModelMode() string
}

// ModelRouter resolves a model mode (e.g. "live_hf", "precomputed",
// "road_damage") to a concrete InferencePort.
type ModelRouter interface {
	Resolve(mode string) (InferencePort, error)
}

// AnalysisStorePort persists analysis results and review records. The MVP uses
// an in-memory implementation; the interface keeps a DB swap cheap.
type AnalysisStorePort interface {
	Save(ctx context.Context, result domain.AnalysisResult) error
	Get(ctx context.Context, analysisID string) (domain.AnalysisResult, error)
	Latest(ctx context.Context) (domain.AnalysisResult, bool)
	List(ctx context.Context) []domain.AnalysisResult
	ApplyReview(ctx context.Context, rec domain.ReviewRecord) (domain.AnalysisResult, error)
}

// MaintenanceReport is the structured output of the OpenRouter reasoner
// (design doc 8). It is explanation-only and never part of the decision path.
type MaintenanceReport struct {
	Summary           string `json:"summary"`
	RecommendedAction string `json:"recommended_action"`
	RiskLevel         string `json:"risk_level"`
	KVKKNote          string `json:"kvkk_note"`
}

// ReasonerPort is the optional LLM explanation layer. The system must work when
// this port is absent or fails (graceful degradation).
type ReasonerPort interface {
	GenerateMaintenanceReport(ctx context.Context, result domain.AnalysisResult) (*MaintenanceReport, error)
}

// EventPublisherPort is an optional async sink for analysis events. The MVP
// uses a no-op; it never blocks or fails the request path.
type EventPublisherPort interface {
	PublishAnalysis(ctx context.Context, result domain.AnalysisResult)
}
