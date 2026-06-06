// Package demo provides precomputed inference adapters. These are NOT mock data:
// they are real model-shaped detection outputs (grounded in COCO/RDD labels and
// refreshed from real HF inference) used as the reliable, offline demo path when
// live HF is slow or no token is present (design doc 9.7, 12).
package demo

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

//go:embed fixtures/street_traffic.json fixtures/road_damage.json
var fixturesFS embed.FS

const precomputedModelID = "civiclens/precomputed-detr-rdd"

// fixtureDetection is the on-disk shape (mirrors the HF object-detection JSON).
type fixtureDetection struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
	Box   struct {
		XMin float64 `json:"xmin"`
		YMin float64 `json:"ymin"`
		XMax float64 `json:"xmax"`
		YMax float64 `json:"ymax"`
	} `json:"box"`
	ModelID string `json:"model_id"`
}

func loadFixture(name string) ([]domain.RawDetection, error) {
	data, err := fixturesFS.ReadFile("fixtures/" + name)
	if err != nil {
		return nil, fmt.Errorf("demo: read fixture %s: %w", name, err)
	}
	var fds []fixtureDetection
	if err := json.Unmarshal(data, &fds); err != nil {
		return nil, fmt.Errorf("demo: parse fixture %s: %w", name, err)
	}
	out := make([]domain.RawDetection, 0, len(fds))
	for _, fd := range fds {
		out = append(out, domain.RawDetection{
			Label:      fd.Label,
			Confidence: fd.Score,
			BBox:       domain.BoundingBox{XMin: fd.Box.XMin, YMin: fd.Box.YMin, XMax: fd.Box.XMax, YMax: fd.Box.YMax},
			ModelID:    fd.ModelID,
		})
	}
	return out, nil
}

// PrecomputedAdapter serves precomputed detections selected by source ref.
type PrecomputedAdapter struct {
	defaultFixture string
}

// NewPrecomputedAdapter builds the adapter (default: street_traffic scene).
func NewPrecomputedAdapter() *PrecomputedAdapter {
	return &PrecomputedAdapter{defaultFixture: "street_traffic.json"}
}

// Detect returns precomputed detections. The fixture is chosen by source ref so
// the demo can show both a traffic scene and the road-damage hero shot.
func (a *PrecomputedAdapter) Detect(_ context.Context, in appvision.InferenceInput) ([]domain.RawDetection, error) {
	return loadFixture(a.fixtureFor(in.SourceRef))
}

func (a *PrecomputedAdapter) fixtureFor(sourceRef string) string {
	ref := strings.ToLower(sourceRef)
	if strings.Contains(ref, "road") || strings.Contains(ref, "damage") || strings.Contains(ref, "pothole") {
		return "road_damage.json"
	}
	return a.defaultFixture
}

// ModelID identifies the precomputed source.
func (a *PrecomputedAdapter) ModelID() string { return precomputedModelID }

// ModelMode reports precomputed transparency mode.
func (a *PrecomputedAdapter) ModelMode() string { return domain.ModelModePrecomputed }
