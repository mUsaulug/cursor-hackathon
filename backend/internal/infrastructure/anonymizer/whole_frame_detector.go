package anonymizer

import (
	"context"
	"image"

	"cursor-hackathon/backend/internal/shared/imaging"
)

// WholeFrameDetector marks the entire frame as PII for irreversible blur when no
// HF-based detector is available. Used for real citizen/staff uploads offline.
type WholeFrameDetector struct{}

// DetectPII returns the full image bounds so the anonymizer pixelates everything.
func (WholeFrameDetector) DetectPII(_ context.Context, img []byte) ([]image.Rectangle, error) {
	src, _, err := imaging.Decode(img)
	if err != nil {
		return nil, err
	}
	return []image.Rectangle{src.Bounds()}, nil
}
