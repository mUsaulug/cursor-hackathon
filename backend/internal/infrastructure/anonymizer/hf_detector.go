package anonymizer

import (
	"context"
	"image"

	appvision "cursor-hackathon/backend/internal/application/vision"
	"cursor-hackathon/backend/internal/infrastructure/huggingface"
	"cursor-hackathon/backend/internal/shared/imaging"
)

// piiLabels are the COCO classes whose regions must be blurred for KVKK: people
// and vehicles (vehicles carry plates). Detection here is ONLY to redact.
var piiLabels = map[string]bool{
	"person":     true,
	"bicycle":    true,
	"motorcycle": true,
	"car":        true,
	"truck":      true,
	"bus":        true,
}

// HFDetector reuses the live DETR adapter to locate people/vehicles, then hands
// their boxes to the blur step. It is opt-in (requires an HF token via the
// underlying client) and never used for identification.
type HFDetector struct {
	adapter *huggingface.DETRAdapter
}

// NewHFDetector wraps a DETR adapter.
func NewHFDetector(adapter *huggingface.DETRAdapter) *HFDetector {
	return &HFDetector{adapter: adapter}
}

// DetectPII returns rectangles for every detected person/vehicle.
func (d *HFDetector) DetectPII(ctx context.Context, img []byte) ([]image.Rectangle, error) {
	raws, err := d.adapter.Detect(ctx, appvision.InferenceInput{Image: img})
	if err != nil {
		return nil, err
	}
	var rects []image.Rectangle
	for _, r := range raws {
		if piiLabels[r.Label] {
			rects = append(rects, imaging.BBoxToRect(r.BBox))
		}
	}
	return rects, nil
}
