// Package verification handles task-completion proof: field staff upload an
// "after" photo, the system anonymizes + analyzes it and compares to the
// "before" problem deterministically (no LLM) to decide whether it is resolved.
package verification

import (
	evidence "cursor-hackathon/backend/internal/domain/evidence"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

// resolvedConfidence is the threshold above which an "after" detection of the
// same problem type is considered to still be present.
const resolvedConfidence = 0.50

// verify compares the before problem type to the after detections and returns a
// deterministic verification outcome.
func verify(beforeType string, after []domain.Detection) evidence.Verification {
	if beforeType == "" || beforeType == domain.TypeUnknown {
		// We cannot reason about an unknown before-state.
		return evidence.VerificationNeedsHuman
	}
	for _, d := range after {
		if d.NormalizedObjectType == beforeType &&
			d.ReviewStatus != domain.ReviewRejected &&
			d.Confidence >= resolvedConfidence {
			// Same problem still detected at the location.
			return evidence.VerificationStillPresent
		}
	}
	// The before problem type is no longer detected -> likely resolved.
	return evidence.VerificationLikelyResolved
}
