package store

import (
	"context"
	"sync"

	evidence "cursor-hackathon/backend/internal/domain/evidence"
)

// EvidenceInMemory is a goroutine-safe in-memory evidence store keyed by task.
type EvidenceInMemory struct {
	mu     sync.RWMutex
	byID   map[string]evidence.CompletionEvidence
	byTask map[string]string // taskID -> evidenceID (latest)
}

// NewEvidenceInMemory builds the store.
func NewEvidenceInMemory() *EvidenceInMemory {
	return &EvidenceInMemory{
		byID:   map[string]evidence.CompletionEvidence{},
		byTask: map[string]string{},
	}
}

// Save stores or replaces evidence.
func (s *EvidenceInMemory) Save(_ context.Context, e evidence.CompletionEvidence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[e.EvidenceID] = e
	s.byTask[e.TaskID] = e.EvidenceID
	return nil
}

// GetByTask returns the latest evidence for a task.
func (s *EvidenceInMemory) GetByTask(_ context.Context, taskID string) (evidence.CompletionEvidence, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byTask[taskID]
	if !ok {
		return evidence.CompletionEvidence{}, false
	}
	return s.byID[id], true
}
