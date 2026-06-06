package huggingface

import (
	"context"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

// DETRModelID is the live COCO baseline (design doc 7.1).
const DETRModelID = "facebook/detr-resnet-50"

// DETRAdapter is the live HF object-detection adapter. It implements
// application/vision.InferencePort. Register it only when a token is present;
// the model router falls back to precomputed otherwise.
type DETRAdapter struct {
	client *Client
}

// NewDETRAdapter builds the adapter over an HF client.
func NewDETRAdapter(client *Client) *DETRAdapter {
	return &DETRAdapter{client: client}
}

// Detect runs live inference and maps the HF response into raw domain
// detections. The CivicLens Core pipeline does all normalization/policy.
func (a *DETRAdapter) Detect(ctx context.Context, in appvision.InferenceInput) ([]domain.RawDetection, error) {
	if len(in.Image) == 0 {
		return nil, domain.ErrNoImage
	}
	resp, err := a.client.detect(ctx, DETRModelID, in.Image)
	if err != nil {
		return nil, err
	}
	out := make([]domain.RawDetection, 0, len(resp))
	for _, d := range resp {
		out = append(out, domain.RawDetection{
			Label:      d.Label,
			Confidence: d.Score,
			BBox: domain.BoundingBox{
				XMin: d.Box.XMin,
				YMin: d.Box.YMin,
				XMax: d.Box.XMax,
				YMax: d.Box.YMax,
			},
			ModelID: DETRModelID,
		})
	}
	return out, nil
}

// ModelID identifies the live model.
func (a *DETRAdapter) ModelID() string { return DETRModelID }

// ModelMode reports the transparency mode shown in the dashboard.
func (a *DETRAdapter) ModelMode() string { return domain.ModelModeLiveHF }
