package intake

import (
	"context"
	"testing"
	"time"

	appvision "cursor-hackathon/backend/internal/application/vision"
	report "cursor-hackathon/backend/internal/domain/report"
	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/config"
)

type passthroughAnonymizer struct{}

func (passthroughAnonymizer) Anonymize(_ context.Context, img []byte) (AnonymizationResult, error) {
	return AnonymizationResult{Image: img, Width: 640, Height: 480, Strategy: domain.PIIStrategyAvoidanceByDesign}, nil
}

type stubAnalyzer struct{ dets []domain.Detection }

func (s stubAnalyzer) Execute(_ context.Context, cmd appvision.AnalyzeCommand) (domain.AnalysisResult, error) {
	return domain.AnalysisResult{
		AnalysisID: "ana_" + cmd.ReportID,
		ReportID:   cmd.ReportID,
		Detections: s.dets,
		KVKKSafe:   true,
	}, nil
}

// in-memory report store for tests (reuse domain semantics).
type memReports struct {
	items []report.Report
}

func (m *memReports) Save(_ context.Context, r report.Report) error {
	m.items = append(m.items, r)
	return nil
}
func (m *memReports) Get(_ context.Context, id string) (report.Report, error) {
	for _, r := range m.items {
		if r.ReportID == id {
			return r, nil
		}
	}
	return report.Report{}, report.ErrReportNotFound
}
func (m *memReports) List(context.Context) []report.Report { return m.items }
func (m *memReports) FindRecentDuplicate(_ context.Context, lat, lng float64, pt string) (report.Report, bool) {
	for _, r := range m.items {
		if r.Status == report.StatusMerged || r.Location == nil {
			continue
		}
		if r.ProblemType == pt && r.Location.Lat == lat && r.Location.Lng == lng {
			return r, true
		}
	}
	return report.Report{}, false
}
func (m *memReports) IncrementDuplicate(_ context.Context, id string) (report.Report, error) {
	for i := range m.items {
		if m.items[i].ReportID == id {
			m.items[i].DuplicateCount++
			return m.items[i], nil
		}
	}
	return report.Report{}, report.ErrReportNotFound
}

func newUC(t *testing.T, dets []domain.Detection, store ReportStorePort) *CreateReportUseCase {
	t.Helper()
	rules, err := config.Load()
	if err != nil {
		t.Fatalf("rules: %v", err)
	}
	n := 0
	return NewCreateReportUseCase(passthroughAnonymizer{}, stubAnalyzer{dets: dets}, store, rules,
		WithClock(func() time.Time { return time.Date(2026, 6, 6, 10, 0, 0, 0, time.UTC) }),
		WithIDFunc(func() string { n++; return "rep_" + string(rune('0'+n)) }),
	)
}

func roadDamageDetection() domain.Detection {
	return domain.Detection{NormalizedObjectType: domain.TypeRoadDamage, Priority: domain.PriorityHigh, ReviewStatus: domain.ReviewAutoAccepted, Confidence: 0.9}
}

func TestIntakeRoutesAndPrioritizes(t *testing.T) {
	store := &memReports{}
	uc := newUC(t, []domain.Detection{roadDamageDetection()}, store)

	rep, err := uc.Execute(context.Background(), CreateReportCommand{
		Image:        []byte{1, 2, 3},
		Description:  "Yolda buyuk cukur var",
		SourceType:   report.SourceCitizenMobile,
		ReporterRole: report.ReporterCitizen,
		Location:     &report.Location{Lat: 41.0, Lng: 29.0},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if rep.ProblemType != domain.TypeRoadDamage {
		t.Errorf("problem_type = %q", rep.ProblemType)
	}
	if rep.Priority != string(domain.PriorityHigh) {
		t.Errorf("priority = %q", rep.Priority)
	}
	if rep.AssignedDept != "Fen Isleri Mudurlugu" {
		t.Errorf("department = %q", rep.AssignedDept)
	}
	if rep.Status != report.StatusWaitingForReview {
		t.Errorf("status = %q", rep.Status)
	}
	if rep.AnalysisID == "" {
		t.Error("analysis_id should be linked")
	}
}

func TestIntakeTextSignalWhenModelUnsure(t *testing.T) {
	store := &memReports{}
	// Model returns nothing -> text signal must classify from the description.
	uc := newUC(t, nil, store)
	rep, err := uc.Execute(context.Background(), CreateReportCommand{
		Image:       []byte{1},
		Description: "Cop konteyneri tasmis",
		Location:    &report.Location{Lat: 1, Lng: 1},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if rep.ProblemType != domain.TypeWasteAsset {
		t.Errorf("text signal problem_type = %q, want waste_asset", rep.ProblemType)
	}
}

func TestIntakeDeduplicates(t *testing.T) {
	store := &memReports{}
	uc := newUC(t, []domain.Detection{roadDamageDetection()}, store)
	loc := &report.Location{Lat: 41.0, Lng: 29.0}

	first, _ := uc.Execute(context.Background(), CreateReportCommand{Image: []byte{1}, Location: loc})
	second, _ := uc.Execute(context.Background(), CreateReportCommand{Image: []byte{1}, Location: loc})

	if second.Status != report.StatusMerged {
		t.Errorf("second report status = %q, want merged", second.Status)
	}
	if second.DuplicateOf != first.ReportID {
		t.Errorf("duplicate_of = %q, want %q", second.DuplicateOf, first.ReportID)
	}
}
