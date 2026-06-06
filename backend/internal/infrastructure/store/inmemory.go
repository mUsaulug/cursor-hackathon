// Package store provides an in-memory implementation of the analysis store.
// MVP persistence is deliberately in-memory (design doc 12: "DB yok ->
// in-memory result"); the AnalysisStorePort keeps a DB swap cheap later.
package store

import (
	"context"
	"sync"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// InMemory is a goroutine-safe analysis store.
type InMemory struct {
	mu      sync.RWMutex
	byID    map[string]domain.AnalysisResult
	order   []string // insertion order; last == latest
	maxKeep int      // cap to bound memory during a long demo
}

// NewInMemory builds the store. keep<=0 means unbounded.
func NewInMemory(keep int) *InMemory {
	return &InMemory{
		byID:    map[string]domain.AnalysisResult{},
		maxKeep: keep,
	}
}

// Save stores (or replaces) an analysis result.
func (s *InMemory) Save(_ context.Context, result domain.AnalysisResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byID[result.AnalysisID]; !exists {
		s.order = append(s.order, result.AnalysisID)
		s.evictIfNeeded()
	}
	s.byID[result.AnalysisID] = result
	return nil
}

// Get returns a stored result by id.
func (s *InMemory) Get(_ context.Context, analysisID string) (domain.AnalysisResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.byID[analysisID]
	if !ok {
		return domain.AnalysisResult{}, domain.ErrAnalysisNotFound
	}
	return r, nil
}

// Latest returns the most recently saved result.
func (s *InMemory) Latest(_ context.Context) (domain.AnalysisResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.order) == 0 {
		return domain.AnalysisResult{}, false
	}
	return s.byID[s.order[len(s.order)-1]], true
}

// List returns all stored results in insertion order.
func (s *InMemory) List(_ context.Context) []domain.AnalysisResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AnalysisResult, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.byID[id])
	}
	return out
}

// ApplyReview applies a human decision to a detection: accepted ->
// auto_accepted, rejected -> rejected. It locates the owning analysis by
// detection id and returns the updated analysis.
func (s *InMemory) ApplyReview(_ context.Context, rec domain.ReviewRecord) (domain.AnalysisResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newStatus := domain.ReviewAutoAccepted
	if rec.Decision == domain.ReviewDecisionRejected {
		newStatus = domain.ReviewRejected
	}

	for _, id := range s.order {
		result := s.byID[id]
		for i := range result.Detections {
			if result.Detections[i].ID == rec.DetectionID {
				result.Detections[i].ReviewStatus = newStatus
				s.byID[id] = result
				return result, nil
			}
		}
	}
	return domain.AnalysisResult{}, domain.ErrDetectionNotFound
}

func (s *InMemory) evictIfNeeded() {
	if s.maxKeep <= 0 || len(s.order) <= s.maxKeep {
		return
	}
	oldest := s.order[0]
	s.order = s.order[1:]
	delete(s.byID, oldest)
}
