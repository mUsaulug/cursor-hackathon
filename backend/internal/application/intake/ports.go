// Package intake orchestrates Wave 2 report ingestion: anonymize (KVKK) ->
// analyze (vision pipeline) -> text signal -> duplicate detection -> department
// routing -> persisted Report. It depends only on domain types and ports; the
// decision steps remain deterministic (LLM is never in the decision path).
package intake

import (
	"context"

	appvision "cursor-hackathon/backend/internal/application/vision"
	report "cursor-hackathon/backend/internal/domain/report"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

// AnonymizerPort blurs PII regions before any inference/persistence (KVKK).
type AnonymizerPort interface {
	Anonymize(ctx context.Context, img []byte) (AnonymizationResult, error)
}

// AnonymizationResult is the anonymizer output (transport-agnostic). Width/Height
// come from the decode the anonymizer already performs, so the application layer
// never needs to import an image codec.
type AnonymizationResult struct {
	Image          []byte
	Width          int
	Height         int
	RegionsBlurred int
	Strategy       string
	Anonymized     bool
}

// AnalyzerPort runs the vision decision chain over an (already anonymized) image.
type AnalyzerPort interface {
	Execute(ctx context.Context, cmd appvision.AnalyzeCommand) (domain.AnalysisResult, error)
}

// ReportStorePort persists reports.
type ReportStorePort interface {
	Save(ctx context.Context, r report.Report) error
	Get(ctx context.Context, id string) (report.Report, error)
	List(ctx context.Context) []report.Report
	// FindRecentDuplicate returns an open report at ~the same location with the
	// same problem type within the dedup window, if any.
	FindRecentDuplicate(ctx context.Context, lat, lng float64, problemType string) (report.Report, bool)
	IncrementDuplicate(ctx context.Context, reportID string) (report.Report, error)
}
