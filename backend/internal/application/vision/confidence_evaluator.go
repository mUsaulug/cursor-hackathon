package vision

import domain "cursor-hackathon/backend/internal/domain/vision"

// evaluateConfidence maps a detection score to a review status using per-type
// auto-accept thresholds and the global needs-review floor (design doc 5.4):
//
//	score >= auto_accept(type)        -> auto_accepted
//	needs_review <= score < accept    -> needs_review
//	score < needs_review              -> rejected
func (p *Pipeline) evaluateConfidence(objectType string, score float64) domain.ReviewStatus {
	autoAccept := p.rules.Confidence.AutoAccept(objectType)
	needsReview := p.rules.Confidence.DefaultNeedsReview

	switch {
	case score >= autoAccept:
		return domain.ReviewAutoAccepted
	case score >= needsReview:
		return domain.ReviewNeedsReview
	default:
		return domain.ReviewRejected
	}
}
