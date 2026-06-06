package vision

import (
	"context"
	"testing"
	"time"

	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/config"
)

// --- test doubles ---

type fakeAdapter struct {
	raws []domain.RawDetection
}

func (f fakeAdapter) Detect(_ context.Context, _ InferenceInput) ([]domain.RawDetection, error) {
	return f.raws, nil
}
func (f fakeAdapter) ModelID() string   { return "fake/model" }
func (f fakeAdapter) ModelMode() string { return domain.ModelModePrecomputed }

type fakeRouter struct{ adapter InferencePort }

func (r fakeRouter) Resolve(string) (InferencePort, error) { return r.adapter, nil }

type fakeStore struct {
	saved   []domain.AnalysisResult
	publish int
}

func (s *fakeStore) Save(_ context.Context, r domain.AnalysisResult) error {
	s.saved = append(s.saved, r)
	return nil
}
func (s *fakeStore) Get(context.Context, string) (domain.AnalysisResult, error) {
	return domain.AnalysisResult{}, domain.ErrAnalysisNotFound
}
func (s *fakeStore) Latest(context.Context) (domain.AnalysisResult, bool) {
	if len(s.saved) == 0 {
		return domain.AnalysisResult{}, false
	}
	return s.saved[len(s.saved)-1], true
}
func (s *fakeStore) List(context.Context) []domain.AnalysisResult { return s.saved }
func (s *fakeStore) ApplyReview(context.Context, domain.ReviewRecord) (domain.AnalysisResult, error) {
	return domain.AnalysisResult{}, nil
}

type fakeEvents struct{ count int }

func (e *fakeEvents) PublishAnalysis(context.Context, domain.AnalysisResult) { e.count++ }

func newUseCase(t *testing.T, raws []domain.RawDetection, store AnalysisStorePort, events EventPublisherPort) *AnalyzeImageUseCase {
	t.Helper()
	rules, err := config.Load()
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	pipe := NewPipeline(rules, WithIDFunc(func() string { return "fixed-det" }))
	router := fakeRouter{adapter: fakeAdapter{raws: raws}}
	opts := []UseCaseOption{
		WithClock(func() time.Time { return time.Date(2026, 6, 6, 10, 30, 0, 0, time.UTC) }),
		WithUseCaseIDFunc(func() string { return "ana-fixed" }),
	}
	if events != nil {
		opts = append(opts, WithEventPublisher(events))
	}
	return NewAnalyzeImageUseCase(router, pipe, store, opts...)
}

func TestExecuteHappyPath(t *testing.T) {
	store := &fakeStore{}
	events := &fakeEvents{}
	uc := newUseCase(t, []domain.RawDetection{
		raw("traffic light", 0.91),
		raw("person", 0.99), // must be filtered by privacy guard
	}, store, events)

	res, err := uc.Execute(context.Background(), AnalyzeCommand{
		Image:       []byte{0xFF, 0xD8, 0xFF}, // non-empty
		SourceType:  domain.SourceTypeUpload,
		SourceRef:   "sample_01",
		ImageWidth:  640,
		ImageHeight: 480,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.SchemaVersion != domain.SchemaVersion {
		t.Errorf("schema_version = %q", res.SchemaVersion)
	}
	if res.AnalysisID != "ana-fixed" {
		t.Errorf("analysis_id = %q", res.AnalysisID)
	}
	if res.RawImageStored {
		t.Error("raw_image_stored must be false")
	}
	if !res.KVKKSafe {
		t.Error("kvkk_safe must be true")
	}
	if res.ImageWidth != 640 || res.ImageHeight != 480 {
		t.Errorf("dimensions = %dx%d", res.ImageWidth, res.ImageHeight)
	}
	if len(res.Detections) != 1 {
		t.Fatalf("detections = %d, want 1 (person filtered)", len(res.Detections))
	}
	if res.Privacy.BlockedCount != 1 {
		t.Errorf("blocked_count = %d, want 1", res.Privacy.BlockedCount)
	}
	if res.CreatedAt != "2026-06-06T10:30:00Z" {
		t.Errorf("created_at = %q", res.CreatedAt)
	}
	if len(store.saved) != 1 {
		t.Errorf("store saved %d results, want 1", len(store.saved))
	}
	if events.count != 1 {
		t.Errorf("event published %d times, want 1", events.count)
	}
}

func TestExecuteNoImageRejected(t *testing.T) {
	store := &fakeStore{}
	uc := newUseCase(t, nil, store, nil)
	_, err := uc.Execute(context.Background(), AnalyzeCommand{
		SourceType: domain.SourceTypeUpload,
	})
	if err != domain.ErrNoImage {
		t.Fatalf("err = %v, want ErrNoImage", err)
	}
}
