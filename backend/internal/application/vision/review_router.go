package vision

import domain "cursor-hackathon/backend/internal/domain/vision"

// route decides the final review status. It combines the confidence evaluator
// with the unknown-type rule: an unmapped/unknown object is never auto-accepted
// even at high confidence — it is downgraded to needs_review so a human can
// confirm an out-of-ontology object (design doc 5.6). Rejected stays rejected.
func (p *Pipeline) route(objectType string, score float64, mapped bool) domain.ReviewStatus {
	status := p.evaluateConfidence(objectType, score)

	if (!mapped || objectType == domain.TypeUnknown) && status == domain.ReviewAutoAccepted {
		return domain.ReviewNeedsReview
	}
	return status
}
