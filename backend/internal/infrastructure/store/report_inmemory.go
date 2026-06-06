package store

import (
	"context"
	"math"
	"sync"
	"time"

	report "cursor-hackathon/backend/internal/domain/report"
)

// ReportInMemory is a goroutine-safe in-memory report store with simple
// location+type+time duplicate detection. The product swaps this for Postgres
// (+PostGIS) behind the same port.
type ReportInMemory struct {
	mu          sync.RWMutex
	byID        map[string]report.Report
	order       []string
	dedupRadius float64       // degrees (~0.0005 ≈ 55m)
	dedupWindow time.Duration // recent window for duplicates
	now         func() time.Time
}

// NewReportInMemory builds the store with sane dedup defaults.
func NewReportInMemory() *ReportInMemory {
	return &ReportInMemory{
		byID:        map[string]report.Report{},
		dedupRadius: 0.0005,
		dedupWindow: 24 * time.Hour,
		now:         time.Now,
	}
}

// Save stores or replaces a report.
func (s *ReportInMemory) Save(_ context.Context, r report.Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[r.ReportID]; !ok {
		s.order = append(s.order, r.ReportID)
	}
	s.byID[r.ReportID] = r
	return nil
}

// Get returns a report by id.
func (s *ReportInMemory) Get(_ context.Context, id string) (report.Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.byID[id]
	if !ok {
		return report.Report{}, report.ErrReportNotFound
	}
	return r, nil
}

// List returns reports in insertion order.
func (s *ReportInMemory) List(_ context.Context) []report.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]report.Report, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.byID[id])
	}
	return out
}

// FindRecentDuplicate returns an open report at ~the same location with the same
// problem type within the dedup window. Merged/rejected reports are skipped.
func (s *ReportInMemory) FindRecentDuplicate(_ context.Context, lat, lng float64, problemType string) (report.Report, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cutoff := s.now().Add(-s.dedupWindow)
	for i := len(s.order) - 1; i >= 0; i-- {
		r := s.byID[s.order[i]]
		if r.Status == report.StatusMerged || r.Status == report.StatusRejected {
			continue
		}
		if r.ProblemType != problemType || r.Location == nil {
			continue
		}
		if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil && t.Before(cutoff) {
			continue
		}
		if math.Abs(r.Location.Lat-lat) <= s.dedupRadius && math.Abs(r.Location.Lng-lng) <= s.dedupRadius {
			return r, true
		}
	}
	return report.Report{}, false
}

// IncrementDuplicate bumps the duplicate counter on an existing report.
func (s *ReportInMemory) IncrementDuplicate(_ context.Context, reportID string) (report.Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.byID[reportID]
	if !ok {
		return report.Report{}, report.ErrReportNotFound
	}
	r.DuplicateCount++
	s.byID[reportID] = r
	return r, nil
}
