// Package anonymizer enforces KVKK at ingest: it detects PII regions (people,
// vehicles → faces/plates) and irreversibly blurs them BEFORE any inference,
// persistence, or transmission. In Wave 2 the data is real citizen/staff
// photos, so this is mandatory (not avoidance-by-design). Detection is used
// ONLY to locate regions to redact — never to identify (design doc 5.3).
package anonymizer

import (
	"context"
	"image"

	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/imaging"
)

// Detector locates PII regions to redact. Implementations may use a model
// (HF person/vehicle detection) or be a no-op for synthetic/demo inputs.
type Detector interface {
	DetectPII(ctx context.Context, img []byte) ([]image.Rectangle, error)
}

// Result is the outcome of anonymization.
type Result struct {
	Image          []byte // anonymized JPEG (raw input must be discarded)
	RegionsBlurred int
	Strategy       string // domain.PIIStrategyBlurApplied | PIIStrategyAvoidanceByDesign
	Anonymized     bool
}

// Anonymizer blurs PII regions using a Detector + stdlib pixelation.
type Anonymizer struct {
	detector Detector
	block    int
	strategy string
}

// New builds an anonymizer. strategy reflects how PII is handled so the privacy
// report can be honest (blur_applied vs avoidance_by_design for no-op/demo).
func New(detector Detector, strategy string) *Anonymizer {
	if strategy == "" {
		strategy = domain.PIIStrategyAvoidanceByDesign
	}
	return &Anonymizer{detector: detector, block: 22, strategy: strategy}
}

// Anonymize decodes (JPEG/PNG), blurs detected PII regions, and re-encodes.
// On a decode error it returns the error so the caller can refuse the image —
// silently storing a raw frame is forbidden.
func (a *Anonymizer) Anonymize(ctx context.Context, img []byte) (Result, error) {
	src, _, err := imaging.Decode(img)
	if err != nil {
		return Result{}, err
	}

	var regions []image.Rectangle
	if a.detector != nil {
		regions, err = a.detector.DetectPII(ctx, img)
		if err != nil {
			// KVKK-safe fallback: if detection fails, blur the whole frame
			// rather than risk leaking PII.
			regions = []image.Rectangle{src.Bounds()}
		}
	}

	if len(regions) == 0 {
		// Nothing to redact (e.g. synthetic/demo image with no PII).
		out, encErr := imaging.EncodeJPEG(src, 88)
		if encErr != nil {
			return Result{}, encErr
		}
		return Result{Image: out, RegionsBlurred: 0, Strategy: a.strategy, Anonymized: false}, nil
	}

	blurred := imaging.PixelateRegions(src, regions, a.block)
	out, err := imaging.EncodeJPEG(blurred, 88)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Image:          out,
		RegionsBlurred: len(regions),
		Strategy:       domain.PIIStrategyBlurApplied,
		Anonymized:     true,
	}, nil
}
