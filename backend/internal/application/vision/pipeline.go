package vision

import (
	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/config"
	"cursor-hackathon/backend/internal/shared/idgen"
)

// Pipeline is the deterministic CivicLens Core decision chain. Every step is
// rule-driven Go code; no step calls an LLM. It is safe for concurrent use:
// it holds only immutable rules and stateless function deps.
type Pipeline struct {
	rules *config.Rules
	newID func() string
}

// NewPipeline builds a pipeline from loaded rules. Options allow injecting a
// deterministic ID function in tests.
func NewPipeline(rules *config.Rules, opts ...PipelineOption) *Pipeline {
	p := &Pipeline{rules: rules, newID: idgen.NewUUID}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// PipelineOption configures a Pipeline.
type PipelineOption func(*Pipeline)

// WithIDFunc overrides the detection ID generator (testing).
func WithIDFunc(fn func() string) PipelineOption {
	return func(p *Pipeline) { p.newID = fn }
}

// ProcessResult is the pipeline output: kept detections plus the privacy report
// summarizing what was filtered for KVKK reasons.
type ProcessResult struct {
	Detections []domain.Detection
	Privacy    domain.PrivacyReport
}

// Process runs the full decision chain over raw detections from one inference
// call. Order matters: privacy guard runs before anything is kept, so blocked
// PII classes never reach confidence/priority scoring or the output.
func (p *Pipeline) Process(raws []domain.RawDetection) ProcessResult {
	kept := make([]domain.Detection, 0, len(raws))
	blocked := 0

	for _, raw := range raws {
		// 1. Privacy Guard (KVKK) — deterministic, runs first.
		switch p.privacyDecision(raw.Label) {
		case privacyBlocked:
			blocked++
			continue
		case privacyHidden:
			continue
		}

		// 2. Normalizer — raw label -> normalized object type.
		objectType, mapped := p.normalize(raw.Label)

		// 3. Confidence Evaluator + 4. Review Router -> review status.
		status := p.route(objectType, raw.Confidence, mapped)

		// Rejected detections are dropped from the actionable output.
		if status == domain.ReviewRejected {
			continue
		}

		// 5. Priority Engine.
		priority := p.assignPriority(objectType)

		kept = append(kept, domain.Detection{
			ID:                   p.newID(),
			Label:                raw.Label,
			NormalizedObjectType: objectType,
			Confidence:           raw.Confidence,
			BBox:                 raw.BBox,
			ModelID:              raw.ModelID,
			ReviewStatus:         status,
			Priority:             priority,
			Reason:               buildReason(objectType, priority, status, mapped),
		})
	}

	return ProcessResult{
		Detections: kept,
		Privacy:    p.buildPrivacyReport(blocked),
	}
}
