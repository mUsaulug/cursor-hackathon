package vision

import (
	"sync"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// Router maps model modes to inference adapters and implements ModelRouter.
// When a requested mode is not registered it falls back to the precomputed
// adapter (design doc 12: documented fallback, never a hard failure). The
// resolved adapter reports its own ModelMode, so transparency stays honest.
type Router struct {
	mu       sync.RWMutex
	adapters map[string]InferencePort
	fallback InferencePort
}

// NewRouter builds a router. The fallback (precomputed) adapter is required so
// resolution always succeeds for a known-good demo path.
func NewRouter(fallback InferencePort) *Router {
	return &Router{
		adapters: map[string]InferencePort{},
		fallback: fallback,
	}
}

// Register adds an adapter for a mode. Safe for startup wiring.
func (r *Router) Register(mode string, adapter InferencePort) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[mode] = adapter
}

// Resolve returns the adapter for a mode, or the fallback when unregistered.
func (r *Router) Resolve(mode string) (InferencePort, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if a, ok := r.adapters[mode]; ok {
		return a, nil
	}
	if r.fallback != nil {
		return r.fallback, nil
	}
	return nil, domain.ErrUnknownModelMode
}

// Modes returns the registered mode names (for /model-info transparency).
func (r *Router) Modes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	modes := make([]string, 0, len(r.adapters))
	for m := range r.adapters {
		modes = append(modes, m)
	}
	return modes
}
