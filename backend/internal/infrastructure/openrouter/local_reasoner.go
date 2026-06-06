// Package openrouter provides the optional explanation/reasoning layer
// (design doc 8). The LLM only writes prose; it never makes KVKK/priority
// decisions. The system must degrade gracefully when no key is configured, so
// a deterministic LocalReasoner is always available as a fallback.
package openrouter

import (
	"context"
	"fmt"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

// LocalReasoner builds a maintenance report from the deterministic analysis
// result using templates only. No network, always available.
type LocalReasoner struct{}

// NewLocalReasoner builds the fallback reasoner.
func NewLocalReasoner() *LocalReasoner { return &LocalReasoner{} }

// GenerateMaintenanceReport summarizes detections by priority into a short
// Turkish municipal report.
func (r *LocalReasoner) GenerateMaintenanceReport(_ context.Context, result domain.AnalysisResult) (*appvision.MaintenanceReport, error) {
	var high, needsReview, total int
	for _, d := range result.Detections {
		total++
		if d.Priority == domain.PriorityHigh || d.Priority == domain.PriorityCritical {
			high++
		}
		if d.ReviewStatus == domain.ReviewNeedsReview {
			needsReview++
		}
	}

	risk := "low"
	switch {
	case high > 0:
		risk = "high"
	case total > 0:
		risk = "medium"
	}

	summary := fmt.Sprintf(
		"Goruntude %d aksiyon alinabilir tespit bulundu; %d yuksek oncelikli, %d insan onayi bekliyor.",
		total, high, needsReview,
	)
	if total == 0 {
		summary = "Goruntude aksiyon gerektiren kentsel nesne tespit edilmedi."
	}

	action := "Tespitler belediye envanterine islenmelidir."
	if high > 0 {
		action = "Yuksek oncelikli yol hasarlari oncelikli bakim kuyruguna alinmalidir."
	}
	if needsReview > 0 {
		action += " Dusuk guvenli tespitler saha personelince dogrulanmalidir."
	}

	return &appvision.MaintenanceReport{
		Summary:           summary,
		RecommendedAction: action,
		RiskLevel:         risk,
		KVKKNote:          "Ham goruntu saklanmamistir; kisisel veri riski tasiyan siniflar filtrelenmistir.",
	}, nil
}
