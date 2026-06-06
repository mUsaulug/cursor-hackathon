package demo

import (
	"context"
	"testing"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

func TestPrecomputedFixturesParse(t *testing.T) {
	a := NewPrecomputedAdapter()
	got, err := a.Detect(context.Background(), appvision.InferenceInput{SourceRef: "sample_traffic"})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected traffic fixture detections")
	}
	if a.ModelMode() != domain.ModelModePrecomputed {
		t.Errorf("mode = %q", a.ModelMode())
	}
	// Fixture must include a PII class so the privacy guard has something to filter.
	var hasPerson bool
	for _, d := range got {
		if d.Label == "person" {
			hasPerson = true
		}
	}
	if !hasPerson {
		t.Error("traffic fixture should contain a person to exercise the privacy guard")
	}
}

func TestPrecomputedSelectsRoadDamageByRef(t *testing.T) {
	a := NewPrecomputedAdapter()
	got, err := a.Detect(context.Background(), appvision.InferenceInput{SourceRef: "sample_road_damage_01"})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	for _, d := range got {
		if d.Label != "D40" && d.Label != "D00" && d.Label != "D10" && d.Label != "D20" {
			t.Errorf("unexpected road-damage label %q", d.Label)
		}
	}
}

func TestRoadDamageAdapter(t *testing.T) {
	a := NewRoadDamageAdapter()
	got, err := a.Detect(context.Background(), appvision.InferenceInput{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected road-damage detections")
	}
	if a.ModelMode() != domain.ModelModeRoadDamage {
		t.Errorf("mode = %q", a.ModelMode())
	}
	if a.ModelID() != RoadDamageModelID {
		t.Errorf("model id = %q", a.ModelID())
	}
}
