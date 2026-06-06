package tasking

import (
	"context"
	"errors"
	"time"

	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/idgen"
)

// ErrInvalidTransition is returned when a task state change is not allowed.
var ErrInvalidTransition = errors.New("tasking: invalid status transition")

// ErrReportNotReviewable is returned when a report cannot become a task.
var ErrReportNotReviewable = errors.New("tasking: report is not waiting for review")

// Service orchestrates task creation and lifecycle.
type Service struct {
	tasks   TaskStorePort
	reports ReportPort
	now     func() time.Time
	newID   func() string
}

// Option configures the service (DI for tests).
type Option func(*Service)

// WithClock overrides the clock.
func WithClock(now func() time.Time) Option { return func(s *Service) { s.now = now } }

// WithIDFunc overrides the id generator.
func WithIDFunc(fn func() string) Option { return func(s *Service) { s.newID = fn } }

// NewService wires the tasking service.
func NewService(tasks TaskStorePort, reports ReportPort, opts ...Option) *Service {
	s := &Service{tasks: tasks, reports: reports, now: time.Now, newID: idgen.NewUUID}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CreateFromReport (operator action) accepts a report and creates a task,
// inheriting priority and department; optionally assigns it immediately.
func (s *Service) CreateFromReport(ctx context.Context, reportID, assignedTo string) (task.Task, error) {
	rep, err := s.reports.Get(ctx, reportID)
	if err != nil {
		return task.Task{}, err
	}
	if rep.Status != report.StatusWaitingForReview {
		return task.Task{}, ErrReportNotReviewable
	}

	now := s.now().Format(time.RFC3339)
	status := task.StatusCreated
	if assignedTo != "" {
		status = task.StatusAssigned
	}
	t := task.Task{
		TaskID:       s.newID(),
		ReportID:     reportID,
		AssignedDept: rep.AssignedDept,
		AssignedTo:   assignedTo,
		Priority:     rep.Priority,
		Status:       status,
		SLA:          slaForPriority(rep.Priority),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.tasks.Save(ctx, t); err != nil {
		return task.Task{}, err
	}

	rep.Status = report.StatusTaskCreated
	rep.UpdatedAt = now
	_ = s.reports.Save(ctx, rep)
	return t, nil
}

// RejectReport (operator action) marks a report rejected (no task).
func (s *Service) RejectReport(ctx context.Context, reportID string) (report.Report, error) {
	rep, err := s.reports.Get(ctx, reportID)
	if err != nil {
		return report.Report{}, err
	}
	rep.Status = report.StatusRejected
	rep.UpdatedAt = s.now().Format(time.RFC3339)
	if err := s.reports.Save(ctx, rep); err != nil {
		return report.Report{}, err
	}
	return rep, nil
}

// Assign moves a task to a team/person (created/assigned -> assigned).
func (s *Service) Assign(ctx context.Context, taskID, assignedTo string) (task.Task, error) {
	return s.transition(ctx, taskID, task.StatusAssigned, func(t *task.Task) { t.AssignedTo = assignedTo })
}

// Start marks a task in progress (assigned -> started).
func (s *Service) Start(ctx context.Context, taskID string) (task.Task, error) {
	return s.transition(ctx, taskID, task.StatusStarted, nil)
}

// Get returns a task by id.
func (s *Service) Get(ctx context.Context, id string) (task.Task, error) { return s.tasks.Get(ctx, id) }

// List returns tasks, optionally filtered by assignee.
func (s *Service) List(ctx context.Context, assignedTo string) []task.Task {
	all := s.tasks.List(ctx)
	if assignedTo == "" {
		return all
	}
	out := make([]task.Task, 0)
	for _, t := range all {
		if t.AssignedTo == assignedTo {
			out = append(out, t)
		}
	}
	return out
}

func (s *Service) transition(ctx context.Context, taskID string, to task.Status, mutate func(*task.Task)) (task.Task, error) {
	t, err := s.tasks.Get(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}
	if !task.CanTransition(t.Status, to) {
		return task.Task{}, ErrInvalidTransition
	}
	t.Status = to
	if mutate != nil {
		mutate(&t)
	}
	t.UpdatedAt = s.now().Format(time.RFC3339)
	if err := s.tasks.Save(ctx, t); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

// slaForPriority maps priority to a target resolution window.
func slaForPriority(priority string) string {
	switch domain.Priority(priority) {
	case domain.PriorityCritical:
		return "24h"
	case domain.PriorityHigh:
		return "48h"
	case domain.PriorityMedium:
		return "96h"
	default:
		return "168h"
	}
}
