// Package tasking turns an accepted report into a municipal work order and
// drives its lifecycle (assign -> start -> ... ). It enforces the task state
// machine from the task domain. Decision-free orchestration over ports.
package tasking

import (
	"context"

	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
)

// TaskStorePort persists tasks.
type TaskStorePort interface {
	Save(ctx context.Context, t task.Task) error
	Get(ctx context.Context, id string) (task.Task, error)
	List(ctx context.Context) []task.Task
}

// ReportPort reads and updates reports during review/task creation.
type ReportPort interface {
	Get(ctx context.Context, id string) (report.Report, error)
	Save(ctx context.Context, r report.Report) error
}
