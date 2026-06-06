package vision

import "testing"

func TestRouterResolveAndFallback(t *testing.T) {
	fallback := &fakeAdapter{}
	live := &fakeAdapter{}
	r := NewRouter(fallback)
	r.Register("live_hf", live)

	got, err := r.Resolve("live_hf")
	if err != nil {
		t.Fatalf("Resolve(live_hf): %v", err)
	}
	if got != InferencePort(live) {
		t.Error("expected registered live adapter")
	}

	// Unknown mode falls back (documented degradation, never a hard failure).
	fb, err := r.Resolve("does_not_exist")
	if err != nil {
		t.Fatalf("Resolve(unknown): %v", err)
	}
	if fb != InferencePort(fallback) {
		t.Error("expected fallback adapter for unknown mode")
	}
}

func TestRouterModes(t *testing.T) {
	r := NewRouter(&fakeAdapter{})
	r.Register("precomputed", &fakeAdapter{})
	r.Register("live_hf", &fakeAdapter{})
	if len(r.Modes()) != 2 {
		t.Errorf("Modes len = %d, want 2", len(r.Modes()))
	}
}
