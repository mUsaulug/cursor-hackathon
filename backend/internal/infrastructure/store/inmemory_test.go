package store

import (
	"context"
	"testing"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

func result(id string, dets ...domain.Detection) domain.AnalysisResult {
	return domain.AnalysisResult{AnalysisID: id, Detections: dets}
}

func TestSaveGetLatest(t *testing.T) {
	s := NewInMemory(0)
	ctx := context.Background()
	_ = s.Save(ctx, result("a"))
	_ = s.Save(ctx, result("b"))

	if _, err := s.Get(ctx, "a"); err != nil {
		t.Errorf("Get(a): %v", err)
	}
	if _, err := s.Get(ctx, "missing"); err != domain.ErrAnalysisNotFound {
		t.Errorf("Get(missing) err = %v, want ErrAnalysisNotFound", err)
	}
	latest, ok := s.Latest(ctx)
	if !ok || latest.AnalysisID != "b" {
		t.Errorf("Latest = %v ok=%v, want b", latest.AnalysisID, ok)
	}
	if len(s.List(ctx)) != 2 {
		t.Errorf("List len = %d, want 2", len(s.List(ctx)))
	}
}

func TestApplyReview(t *testing.T) {
	s := NewInMemory(0)
	ctx := context.Background()
	_ = s.Save(ctx, result("a", domain.Detection{ID: "d1", ReviewStatus: domain.ReviewNeedsReview}))

	updated, err := s.ApplyReview(ctx, domain.ReviewRecord{DetectionID: "d1", Decision: domain.ReviewDecisionAccepted})
	if err != nil {
		t.Fatalf("ApplyReview: %v", err)
	}
	if updated.Detections[0].ReviewStatus != domain.ReviewAutoAccepted {
		t.Errorf("status = %q, want auto_accepted", updated.Detections[0].ReviewStatus)
	}

	if _, err := s.ApplyReview(ctx, domain.ReviewRecord{DetectionID: "nope"}); err != domain.ErrDetectionNotFound {
		t.Errorf("err = %v, want ErrDetectionNotFound", err)
	}
}

func TestEviction(t *testing.T) {
	s := NewInMemory(2)
	ctx := context.Background()
	_ = s.Save(ctx, result("a"))
	_ = s.Save(ctx, result("b"))
	_ = s.Save(ctx, result("c"))

	if _, err := s.Get(ctx, "a"); err != domain.ErrAnalysisNotFound {
		t.Errorf("oldest should be evicted, got err=%v", err)
	}
	if len(s.List(ctx)) != 2 {
		t.Errorf("List len = %d, want 2", len(s.List(ctx)))
	}
}
