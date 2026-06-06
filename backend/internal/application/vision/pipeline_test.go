package vision

import (
	"testing"

	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/config"
)

func testPipeline(t *testing.T) *Pipeline {
	t.Helper()
	rules, err := config.Load()
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	n := 0
	return NewPipeline(rules, WithIDFunc(func() string {
		n++
		return "det-" + string(rune('0'+n))
	}))
}

func raw(label string, conf float64) domain.RawDetection {
	return domain.RawDetection{
		Label:      label,
		Confidence: conf,
		BBox:       domain.BoundingBox{XMin: 1, YMin: 2, XMax: 3, YMax: 4},
		ModelID:    "test-model",
	}
}

func TestPrivacyGuardBlocksPII(t *testing.T) {
	p := testPipeline(t)
	out := p.Process([]domain.RawDetection{
		raw("person", 0.99),
		raw("bicycle", 0.95),
		raw("motorcycle", 0.95),
		raw("traffic light", 0.95),
	})
	if len(out.Detections) != 1 {
		t.Fatalf("kept %d detections, want 1 (only traffic light)", len(out.Detections))
	}
	if out.Detections[0].NormalizedObjectType != domain.TypeTrafficSignal {
		t.Errorf("kept type = %q", out.Detections[0].NormalizedObjectType)
	}
	if out.Privacy.BlockedCount != 3 {
		t.Errorf("blocked_count = %d, want 3", out.Privacy.BlockedCount)
	}
	if !out.Privacy.KVKKSafe || out.Privacy.RawImageStored {
		t.Errorf("privacy flags wrong: %+v", out.Privacy)
	}
	if out.Privacy.PIIStrategy != domain.PIIStrategyAvoidanceByDesign {
		t.Errorf("pii_strategy = %q", out.Privacy.PIIStrategy)
	}
}

func TestPrivacyGuardHidesVehicles(t *testing.T) {
	p := testPipeline(t)
	out := p.Process([]domain.RawDetection{
		raw("car", 0.99),
		raw("truck", 0.99),
		raw("bus", 0.99),
	})
	if len(out.Detections) != 0 {
		t.Fatalf("kept %d, want 0 (vehicles hidden)", len(out.Detections))
	}
	// Hidden vehicles are not counted as blocked PII.
	if out.Privacy.BlockedCount != 0 {
		t.Errorf("blocked_count = %d, want 0", out.Privacy.BlockedCount)
	}
}

func TestConfidenceRoutingPerType(t *testing.T) {
	p := testPipeline(t)
	// road_damage auto_accept = 0.85.
	cases := []struct {
		label string
		conf  float64
		want  domain.ReviewStatus
	}{
		{"D40", 0.90, domain.ReviewAutoAccepted},
		{"D40", 0.70, domain.ReviewNeedsReview},
		{"D40", 0.30, domain.ReviewRejected},               // dropped from output
		{"traffic light", 0.78, domain.ReviewAutoAccepted}, // traffic_signal accept=0.75
	}
	for _, c := range cases {
		out := p.Process([]domain.RawDetection{raw(c.label, c.conf)})
		if c.want == domain.ReviewRejected {
			if len(out.Detections) != 0 {
				t.Errorf("%s@%.2f: expected rejected (dropped), got %d kept", c.label, c.conf, len(out.Detections))
			}
			continue
		}
		if len(out.Detections) != 1 {
			t.Fatalf("%s@%.2f: kept %d, want 1", c.label, c.conf, len(out.Detections))
		}
		if got := out.Detections[0].ReviewStatus; got != c.want {
			t.Errorf("%s@%.2f: status = %q, want %q", c.label, c.conf, got, c.want)
		}
	}
}

func TestUnknownNeverAutoAccepted(t *testing.T) {
	p := testPipeline(t)
	out := p.Process([]domain.RawDetection{raw("spaceship", 0.99)})
	if len(out.Detections) != 1 {
		t.Fatalf("kept %d, want 1", len(out.Detections))
	}
	d := out.Detections[0]
	if d.NormalizedObjectType != domain.TypeUnknown {
		t.Errorf("type = %q, want unknown", d.NormalizedObjectType)
	}
	// Unknown at 0.99 must be downgraded to needs_review.
	if d.ReviewStatus != domain.ReviewNeedsReview {
		t.Errorf("status = %q, want needs_review", d.ReviewStatus)
	}
	// And priority must be a real priority, never "needs_review".
	if d.Priority != domain.PriorityLow {
		t.Errorf("priority = %q, want low", d.Priority)
	}
}

func TestPriorityAssignment(t *testing.T) {
	p := testPipeline(t)
	out := p.Process([]domain.RawDetection{
		raw("D40", 0.95),           // road_damage -> high
		raw("traffic light", 0.95), // traffic_signal -> medium
		raw("bench", 0.95),         // street_furniture -> low
	})
	got := map[string]domain.Priority{}
	for _, d := range out.Detections {
		got[d.NormalizedObjectType] = d.Priority
	}
	if got[domain.TypeRoadDamage] != domain.PriorityHigh {
		t.Errorf("road_damage priority = %q, want high", got[domain.TypeRoadDamage])
	}
	if got[domain.TypeTrafficSignal] != domain.PriorityMedium {
		t.Errorf("traffic_signal priority = %q, want medium", got[domain.TypeTrafficSignal])
	}
	if got[domain.TypeStreetFurniture] != domain.PriorityLow {
		t.Errorf("street_furniture priority = %q, want low", got[domain.TypeStreetFurniture])
	}
}

func TestReasonIsNonEmpty(t *testing.T) {
	p := testPipeline(t)
	out := p.Process([]domain.RawDetection{raw("D40", 0.95)})
	if len(out.Detections) != 1 || out.Detections[0].Reason == "" {
		t.Errorf("expected non-empty reason, got %+v", out.Detections)
	}
}
