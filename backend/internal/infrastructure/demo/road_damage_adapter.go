package demo

import (
	"context"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

// RoadDamageModelID is the RDD2022 model whose precomputed outputs power the
// road-damage demo hero (design doc 7.1). DETR (COCO) cannot detect road damage,
// so this dedicated mode always serves the road-damage fixture.
const RoadDamageModelID = "rezzzq/yolo12s-road-damage-rdd2022"

// RoadDamageAdapter always returns the precomputed road-damage detections,
// reporting the RDD model id and the road_damage mode for transparency.
type RoadDamageAdapter struct{}

// NewRoadDamageAdapter builds the adapter.
func NewRoadDamageAdapter() *RoadDamageAdapter { return &RoadDamageAdapter{} }

// Detect returns the road-damage fixture regardless of source ref.
func (a *RoadDamageAdapter) Detect(_ context.Context, _ appvision.InferenceInput) ([]domain.RawDetection, error) {
	return loadFixture("road_damage.json")
}

// ModelID identifies the RDD model.
func (a *RoadDamageAdapter) ModelID() string { return RoadDamageModelID }

// ModelMode reports the road-damage transparency mode.
func (a *RoadDamageAdapter) ModelMode() string { return domain.ModelModeRoadDamage }
