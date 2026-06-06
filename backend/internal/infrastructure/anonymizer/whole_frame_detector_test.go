package anonymizer

import (
	"context"
	"testing"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

func TestWholeFrameDetectorBlursEntireImage(t *testing.T) {
	a := New(WholeFrameDetector{}, domain.PIIStrategyBlurApplied)
	res, err := a.Anonymize(context.Background(), syntheticJPEG(t))
	if err != nil {
		t.Fatalf("Anonymize: %v", err)
	}
	if res.RegionsBlurred != 1 || !res.Anonymized {
		t.Errorf("expected whole-frame blur; got %+v", res)
	}
}
