package tasking

import (
	"context"
	"testing"
	"time"

	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
)

type stubReports struct {
	byID map[string]report.Report
}

func (s *stubReports) Get(_ context.Context, id string) (report.Report, error) {
	r, ok := s.byID[id]
	if !ok {
		return report.Report{}, report.ErrReportNotFound
	}
	return r, nil
}

func (s *stubReports) Save(_ context.Context, r report.Report) error {
	s.byID[r.ReportID] = r
	return nil
}

type stubTasks struct {
	saved []task.Task
}

func (s *stubTasks) Save(_ context.Context, t task.Task) error {
	s.saved = append(s.saved, t)
	return nil
}

func (s *stubTasks) Get(_ context.Context, id string) (task.Task, error) {
	for _, t := range s.saved {
		if t.TaskID == id {
			return t, nil
		}
	}
	return task.Task{}, task.ErrTaskNotFound
}

func (s *stubTasks) List(_ context.Context) []task.Task { return s.saved }

func TestCreateFromReportRejectsNonWaitingStatus(t *testing.T) {
	reports := &stubReports{byID: map[string]report.Report{
		"r1": {ReportID: "r1", Status: report.StatusTaskCreated, Priority: "medium"},
	}}
	svc := NewService(&stubTasks{}, reports, WithClock(func() time.Time { return time.Unix(0, 0) }), WithIDFunc(func() string { return "t1" }))

	_, err := svc.CreateFromReport(context.Background(), "r1", "team")
	if err != ErrReportNotReviewable {
		t.Fatalf("expected ErrReportNotReviewable, got %v", err)
	}
}

func TestCreateFromReportAcceptsWaitingForReview(t *testing.T) {
	reports := &stubReports{byID: map[string]report.Report{
		"r1": {ReportID: "r1", Status: report.StatusWaitingForReview, Priority: "medium", AssignedDept: "roads"},
	}}
	tasks := &stubTasks{}
	svc := NewService(tasks, reports, WithClock(func() time.Time { return time.Unix(0, 0) }), WithIDFunc(func() string { return "t1" }))

	got, err := svc.CreateFromReport(context.Background(), "r1", "team")
	if err != nil {
		t.Fatalf("CreateFromReport: %v", err)
	}
	if got.TaskID != "t1" || got.ReportID != "r1" {
		t.Fatalf("unexpected task: %+v", got)
	}
}
