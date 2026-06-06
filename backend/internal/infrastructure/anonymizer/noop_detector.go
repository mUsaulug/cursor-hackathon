package anonymizer

import (
	"context"
	"image"
)

// NoopDetector finds no PII regions. Used for synthetic/demo inputs that contain
// no people or vehicles, where the strategy is avoidance-by-design. It must NOT
// be used for real citizen/staff photos.
type NoopDetector struct{}

// NewNoopDetector builds a no-op detector.
func NewNoopDetector() *NoopDetector { return &NoopDetector{} }

// DetectPII always returns no regions.
func (NoopDetector) DetectPII(context.Context, []byte) ([]image.Rectangle, error) {
	return nil, nil
}
