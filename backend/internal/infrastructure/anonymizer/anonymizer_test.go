package anonymizer

import (
	"context"
	"errors"
	"image"
	"image/color"
	"testing"

	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/imaging"
)

func syntheticJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 4), uint8(y * 4), 64, 255})
		}
	}
	b, err := imaging.EncodeJPEG(img, 90)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return b
}

type fixedDetector struct{ rects []image.Rectangle }

func (d fixedDetector) DetectPII(context.Context, []byte) ([]image.Rectangle, error) {
	return d.rects, nil
}

type failingDetector struct{}

func (failingDetector) DetectPII(context.Context, []byte) ([]image.Rectangle, error) {
	return nil, errors.New("detector down")
}

func TestAnonymizeBlursDetectedRegions(t *testing.T) {
	a := New(fixedDetector{rects: []image.Rectangle{image.Rect(8, 8, 40, 40)}}, domain.PIIStrategyBlurApplied)
	res, err := a.Anonymize(context.Background(), syntheticJPEG(t))
	if err != nil {
		t.Fatalf("Anonymize: %v", err)
	}
	if res.RegionsBlurred != 1 || !res.Anonymized {
		t.Errorf("expected 1 blurred region, anonymized; got %+v", res)
	}
	if res.Strategy != domain.PIIStrategyBlurApplied {
		t.Errorf("strategy = %q", res.Strategy)
	}
	if _, _, err := imaging.Decode(res.Image); err != nil {
		t.Errorf("output not decodable: %v", err)
	}
}

func TestAnonymizeNoPII(t *testing.T) {
	a := New(NewNoopDetector(), domain.PIIStrategyAvoidanceByDesign)
	res, err := a.Anonymize(context.Background(), syntheticJPEG(t))
	if err != nil {
		t.Fatalf("Anonymize: %v", err)
	}
	if res.RegionsBlurred != 0 || res.Anonymized {
		t.Errorf("expected no blur; got %+v", res)
	}
	if res.Strategy != domain.PIIStrategyAvoidanceByDesign {
		t.Errorf("strategy = %q", res.Strategy)
	}
}

func TestAnonymizeDetectorFailureBlursWholeFrame(t *testing.T) {
	a := New(failingDetector{}, domain.PIIStrategyBlurApplied)
	res, err := a.Anonymize(context.Background(), syntheticJPEG(t))
	if err != nil {
		t.Fatalf("Anonymize: %v", err)
	}
	// KVKK-safe fallback: detector failure must still anonymize (whole frame).
	if res.RegionsBlurred != 1 || !res.Anonymized {
		t.Errorf("expected whole-frame blur on detector failure; got %+v", res)
	}
}

func TestAnonymizeRejectsUndecodable(t *testing.T) {
	a := New(NewNoopDetector(), domain.PIIStrategyAvoidanceByDesign)
	if _, err := a.Anonymize(context.Background(), []byte("not an image")); err == nil {
		t.Fatal("expected error for undecodable input (must not silently pass raw)")
	}
}
