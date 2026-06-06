package verification

import (
	"context"
	"time"

	intake "cursor-hackathon/backend/internal/application/intake"
	appvision "cursor-hackathon/backend/internal/application/vision"
	evidence "cursor-hackathon/backend/internal/domain/evidence"
	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
	"cursor-hackathon/backend/internal/shared/idgen"
)

// EvidenceStorePort persists completion evidence.
type EvidenceStorePort interface {
	Save(ctx context.Context, e evidence.CompletionEvidence) error
	GetByTask(ctx context.Context, taskID string) (evidence.CompletionEvidence, bool)
}

// TaskPort reads/updates tasks.
type TaskPort interface {
	Get(ctx context.Context, id string) (task.Task, error)
	Save(ctx context.Context, t task.Task) error
}

// ReportPort reads reports (for the before problem type/analysis).
type ReportPort interface {
	Get(ctx context.Context, id string) (report.Report, error)
}

// Service orchestrates evidence upload + completion verification.
type Service struct {
	anon     intake.AnonymizerPort
	analyzer intake.AnalyzerPort
	evidence EvidenceStorePort
	tasks    TaskPort
	reports  ReportPort
	now      func() time.Time
	newID    func() string
}

// Option configures the service.
type Option func(*Service)

// WithClock overrides the clock.
func WithClock(now func() time.Time) Option { return func(s *Service) { s.now = now } }

// WithIDFunc overrides the id generator.
func WithIDFunc(fn func() string) Option { return func(s *Service) { s.newID = fn } }

// NewService wires the verification service.
func NewService(anon intake.AnonymizerPort, analyzer intake.AnalyzerPort, ev EvidenceStorePort, tasks TaskPort, reports ReportPort, opts ...Option) *Service {
	s := &Service{anon: anon, analyzer: analyzer, evidence: ev, tasks: tasks, reports: reports, now: time.Now, newID: idgen.NewUUID}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// UploadEvidence (field staff) anonymizes + analyzes the "after" photo, compares
// it to the before problem, records evidence, and advances the task to
// ai_verified. KVKK: anonymize-before-analyze, raw never persisted.
func (s *Service) UploadEvidence(ctx context.Context, taskID string, image []byte, uploadedBy string) (evidence.CompletionEvidence, error) {
	t, err := s.tasks.Get(ctx, taskID)
	if err != nil {
		return evidence.CompletionEvidence{}, err
	}
	if !task.CanTransition(t.Status, task.StatusEvidenceUploaded) {
		return evidence.CompletionEvidence{}, ErrInvalidState
	}

	rep, err := s.reports.Get(ctx, t.ReportID)
	if err != nil {
		return evidence.CompletionEvidence{}, err
	}

	anon, err := s.anon.Anonymize(ctx, image)
	if err != nil {
		return evidence.CompletionEvidence{}, err
	}

	evID := s.newID()
	after, err := s.analyzer.Execute(ctx, appvision.AnalyzeCommand{
		Image:       anon.Image,
		SourceType:  "completion_evidence",
		SourceRef:   evID,
		ReportID:    t.ReportID,
		ImageWidth:  anon.Width,
		ImageHeight: anon.Height,
	})
	if err != nil {
		return evidence.CompletionEvidence{}, err
	}

	outcome := verify(rep.ProblemType, after.Detections)
	now := s.now().Format(time.RFC3339)

	ev := evidence.CompletionEvidence{
		EvidenceID:       evID,
		TaskID:           taskID,
		BeforeAnalysisID: rep.AnalysisID,
		AfterAnalysisID:  after.AnalysisID,
		ImageRef:         "anon_" + evID,
		AIVerification:   outcome,
		ManagerApproval:  evidence.ApprovalPending,
		UploadedBy:       uploadedBy,
		CreatedAt:        now,
	}
	if err := s.evidence.Save(ctx, ev); err != nil {
		return evidence.CompletionEvidence{}, err
	}

	// Advance task: started -> evidence_uploaded -> ai_verified.
	t.Status = task.StatusEvidenceUploaded
	t.UpdatedAt = now
	_ = s.tasks.Save(ctx, t)
	t.Status = task.StatusAIVerified
	_ = s.tasks.Save(ctx, t)

	return ev, nil
}

// CloseTask (manager) approves or reopens a verified task. approved -> completed,
// rejected -> reopened.
func (s *Service) CloseTask(ctx context.Context, taskID, decision string) (task.Task, error) {
	t, err := s.tasks.Get(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	to := task.StatusCompleted
	approval := evidence.ApprovalApproved
	if decision == "reopened" || decision == "rejected" {
		to = task.StatusReopened
		approval = evidence.ApprovalRejected
	}
	if !task.CanTransition(t.Status, to) {
		return task.Task{}, ErrInvalidState
	}

	if ev, ok := s.evidence.GetByTask(ctx, taskID); ok {
		ev.ManagerApproval = approval
		_ = s.evidence.Save(ctx, ev)
	}

	t.Status = to
	t.UpdatedAt = s.now().Format(time.RFC3339)
	if err := s.tasks.Save(ctx, t); err != nil {
		return task.Task{}, err
	}
	return t, nil
}
