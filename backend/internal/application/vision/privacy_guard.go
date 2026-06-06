package vision

import domain "cursor-hackathon/backend/internal/domain/vision"

// privacyDecision classifies a raw label against the KVKK label policy. This is
// deterministic and runs before any detection is kept (design doc 5.3, 9.4).
type privacyDecision int

const (
	privacyKeep    privacyDecision = iota
	privacyBlocked                 // PII class (person/bicycle/motorcycle) — removed + counted
	privacyHidden                  // tracking/plate risk (car/truck/bus) — removed silently
)

func (p *Pipeline) privacyDecision(label string) privacyDecision {
	if p.rules.Ontology.IsBlocked(label) {
		return privacyBlocked
	}
	if p.rules.Ontology.IsHidden(label) {
		return privacyHidden
	}
	return privacyKeep
}

// buildPrivacyReport assembles the KVKK report. In the MVP no raw image is ever
// stored and detection targets are inanimate, so the strategy is
// avoidance-by-design and kvkk_safe is always true once the guard has run.
func (p *Pipeline) buildPrivacyReport(blockedCount int) domain.PrivacyReport {
	return domain.PrivacyReport{
		KVKKSafe:       true,
		RawImageStored: false,
		Anonymized:     false,
		DeletionStatus: domain.DeletionRawNotPersisted,
		BlockedCount:   blockedCount,
		PIIStrategy:    domain.PIIStrategyAvoidanceByDesign,
	}
}
