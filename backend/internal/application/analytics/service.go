// Package analytics produces manager dashboard aggregations over reports and
// tasks. Read-only; deterministic counting (no LLM). Backed by the same stores
// (in-memory now, Postgres later) via list ports.
package analytics

import (
	"context"
	"time"

	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
)

// ReportLister lists reports.
type ReportLister interface {
	List(ctx context.Context) []report.Report
}

// TaskLister lists tasks.
type TaskLister interface {
	List(ctx context.Context) []task.Task
}

// Summary is the manager dashboard payload.
type Summary struct {
	TotalReports        int            `json:"total_reports"`
	ReportsByStatus     map[string]int `json:"reports_by_status"`
	ReportsByType       map[string]int `json:"reports_by_type"`
	ReportsByDepartment map[string]int `json:"reports_by_department"`
	NeedsReview         int            `json:"needs_review"`
	TotalTasks          int            `json:"total_tasks"`
	TasksByStatus       map[string]int `json:"tasks_by_status"`
	CompletedTasks      int            `json:"completed_tasks"`
	AvgResolutionHours  float64        `json:"avg_resolution_hours"`
}

// Service computes dashboard summaries.
type Service struct {
	reports ReportLister
	tasks   TaskLister
}

// NewService wires the analytics service.
func NewService(reports ReportLister, tasks TaskLister) *Service {
	return &Service{reports: reports, tasks: tasks}
}

// Summary aggregates current reports and tasks.
func (s *Service) Summary(ctx context.Context) Summary {
	out := Summary{
		ReportsByStatus:     map[string]int{},
		ReportsByType:       map[string]int{},
		ReportsByDepartment: map[string]int{},
		TasksByStatus:       map[string]int{},
	}

	for _, r := range s.reports.List(ctx) {
		out.TotalReports++
		out.ReportsByStatus[string(r.Status)]++
		if r.ProblemType != "" {
			out.ReportsByType[r.ProblemType]++
		}
		if r.AssignedDept != "" {
			out.ReportsByDepartment[r.AssignedDept]++
		}
		if r.ReviewStatus == "needs_review" {
			out.NeedsReview++
		}
	}

	var totalHours float64
	var resolved int
	for _, t := range s.tasks.List(ctx) {
		out.TotalTasks++
		out.TasksByStatus[string(t.Status)]++
		if t.Status == task.StatusCompleted {
			out.CompletedTasks++
			if h, ok := resolutionHours(t.CreatedAt, t.UpdatedAt); ok {
				totalHours += h
				resolved++
			}
		}
	}
	if resolved > 0 {
		out.AvgResolutionHours = round1(totalHours / float64(resolved))
	}
	return out
}

func resolutionHours(createdAt, updatedAt string) (float64, bool) {
	c, err1 := time.Parse(time.RFC3339, createdAt)
	u, err2 := time.Parse(time.RFC3339, updatedAt)
	if err1 != nil || err2 != nil || !u.After(c) {
		return 0, false
	}
	return u.Sub(c).Hours(), true
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
