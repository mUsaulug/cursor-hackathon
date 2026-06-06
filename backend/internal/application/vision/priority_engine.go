package vision

import domain "cursor-hackathon/backend/internal/domain/vision"

// assignPriority maps a normalized object type to a maintenance priority via the
// priority policy rules (design doc 5.5). Priority is independent of review
// status; an "unknown" object is low priority but still needs review.
func (p *Pipeline) assignPriority(objectType string) domain.Priority {
	return domain.Priority(p.rules.Priority.Priority(objectType))
}
