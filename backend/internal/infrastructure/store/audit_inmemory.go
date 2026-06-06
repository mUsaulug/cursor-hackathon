package store

import (
	"context"
	"sync"

	audit "cursor-hackathon/backend/internal/domain/audit"
)

// AuditInMemory is a goroutine-safe, bounded in-memory audit log.
type AuditInMemory struct {
	mu      sync.RWMutex
	entries []audit.Entry
	maxKeep int
}

// NewAuditInMemory builds the audit log (keeps the most recent maxKeep entries).
func NewAuditInMemory(maxKeep int) *AuditInMemory {
	if maxKeep <= 0 {
		maxKeep = 1000
	}
	return &AuditInMemory{maxKeep: maxKeep}
}

// Record appends an audit entry.
func (s *AuditInMemory) Record(_ context.Context, e audit.Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
	if len(s.entries) > s.maxKeep {
		s.entries = s.entries[len(s.entries)-s.maxKeep:]
	}
}

// List returns audit entries (oldest first).
func (s *AuditInMemory) List(_ context.Context) []audit.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]audit.Entry, len(s.entries))
	copy(out, s.entries)
	return out
}
