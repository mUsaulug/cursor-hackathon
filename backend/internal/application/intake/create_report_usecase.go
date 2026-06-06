package intake

import (
	"context"
	"time"

	appvision "cursor-hackathon/backend/internal/application/vision"
	report "cursor-hackathon/backend/internal/domain/report"
	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/config"
	"cursor-hackathon/backend/internal/shared/idgen"
)

// CreateReportCommand is the intake input from a citizen or staff member.
type CreateReportCommand struct {
	Image        []byte
	Description  string
	SourceType   string
	ReporterRole string
	ReporterID   string
	Location     *report.Location
}

// CreateReportUseCase runs the deterministic intake chain.
type CreateReportUseCase struct {
	anonymizer AnonymizerPort
	analyzer   AnalyzerPort
	reports    ReportStorePort
	rules      *config.Rules
	now        func() time.Time
	newID      func() string
}

// Option configures the use case (DI for tests).
type Option func(*CreateReportUseCase)

// WithClock overrides the timestamp source.
func WithClock(now func() time.Time) Option {
	return func(u *CreateReportUseCase) { u.now = now }
}

// WithIDFunc overrides the report ID generator.
func WithIDFunc(fn func() string) Option {
	return func(u *CreateReportUseCase) { u.newID = fn }
}

// NewCreateReportUseCase wires the intake use case.
func NewCreateReportUseCase(anon AnonymizerPort, analyzer AnalyzerPort, reports ReportStorePort, rules *config.Rules, opts ...Option) *CreateReportUseCase {
	u := &CreateReportUseCase{
		anonymizer: anon,
		analyzer:   analyzer,
		reports:    reports,
		rules:      rules,
		now:        time.Now,
		newID:      idgen.NewUUID,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// Execute ingests one report. Order is KVKK-critical: anonymize first, then
// analyze the anonymized image, then classify/dedup/route deterministically.
func (u *CreateReportUseCase) Execute(ctx context.Context, cmd CreateReportCommand) (report.Report, error) {
	if len(cmd.Image) == 0 {
		return report.Report{}, domain.ErrNoImage
	}

	// 1. KVKK anonymize (blur PII before anything else). A decode failure means
	// we refuse the image rather than risk leaking raw PII.
	anon, err := u.anonymizer.Anonymize(ctx, cmd.Image)
	if err != nil {
		return report.Report{}, err
	}

	reportID := u.newID()
	now := u.now().Format(time.RFC3339)

	// 2. Vision decision chain over the anonymized image.
	var loc *domain.Location
	if cmd.Location != nil {
		loc = &domain.Location{Lat: cmd.Location.Lat, Lng: cmd.Location.Lng}
	}
	analysis, err := u.analyzer.Execute(ctx, appvision.AnalyzeCommand{
		Image:       anon.Image,
		SourceType:  cmd.SourceType,
		SourceRef:   reportID,
		ReportID:    reportID,
		ImageWidth:  anon.Width,
		ImageHeight: anon.Height,
		Location:    loc,
		AnonymizationMeta: &appvision.AnonymizationMeta{
			Strategy:       anon.Strategy,
			Anonymized:     anon.Anonymized,
			RegionsBlurred: anon.RegionsBlurred,
		},
	})
	if err != nil {
		return report.Report{}, err
	}

	// 3. Determine the dominant problem from detections; fall back to the text
	// signal only when the model is unsure.
	problemType, priority, reviewStatus := dominant(analysis.Detections)
	if problemType == "" || problemType == domain.TypeUnknown {
		if signal := classifyText(cmd.Description); signal != "" {
			problemType = signal
		}
	}
	if problemType == "" {
		problemType = domain.TypeUnknown
	}
	if reviewStatus == "" {
		reviewStatus = string(domain.ReviewNeedsReview)
	}

	// 4. Duplicate detection (same location + type within window).
	if cmd.Location != nil {
		if existing, found := u.reports.FindRecentDuplicate(ctx, cmd.Location.Lat, cmd.Location.Lng, problemType); found {
			if _, incErr := u.reports.IncrementDuplicate(ctx, existing.ReportID); incErr == nil {
				merged := u.buildReport(reportID, cmd, analysis.AnalysisID, problemType, priority, reviewStatus, now)
				merged.Status = report.StatusMerged
				merged.DuplicateOf = existing.ReportID
				_ = u.reports.Save(ctx, merged)
				return merged, nil
			}
		}
	}

	// 5. Build, route, persist.
	r := u.buildReport(reportID, cmd, analysis.AnalysisID, problemType, priority, reviewStatus, now)
	if err := u.reports.Save(ctx, r); err != nil {
		return report.Report{}, err
	}
	return r, nil
}

func (u *CreateReportUseCase) buildReport(reportID string, cmd CreateReportCommand, analysisID, problemType, priority, reviewStatus, now string) report.Report {
	return report.Report{
		ReportID:     reportID,
		SourceType:   cmd.SourceType,
		ReporterRole: cmd.ReporterRole,
		ReporterID:   cmd.ReporterID,
		Description:  cmd.Description,
		Location:     cmd.Location,
		ImageRef:     "anon_" + reportID, // anonymized derivative reference only
		AnalysisID:   analysisID,
		ProblemType:  problemType,
		Priority:     priority,
		ReviewStatus: reviewStatus,
		AssignedDept: u.rules.Routing.Department(problemType),
		Status:       report.StatusWaitingForReview,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// dominant picks the most important detection (highest priority, then highest
// confidence) and returns its type/priority/review status.
func dominant(dets []domain.Detection) (string, string, string) {
	var best *domain.Detection
	for i := range dets {
		d := &dets[i]
		if best == nil || rank(d.Priority) > rank(best.Priority) ||
			(rank(d.Priority) == rank(best.Priority) && d.Confidence > best.Confidence) {
			best = d
		}
	}
	if best == nil {
		return "", "", ""
	}
	return best.NormalizedObjectType, string(best.Priority), string(best.ReviewStatus)
}

func rank(p domain.Priority) int {
	switch p {
	case domain.PriorityCritical:
		return 4
	case domain.PriorityHigh:
		return 3
	case domain.PriorityMedium:
		return 2
	case domain.PriorityLow:
		return 1
	default:
		return 0
	}
}
