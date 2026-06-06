package report

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	intake "cursor-hackathon/backend/internal/application/intake"
	appvision "cursor-hackathon/backend/internal/application/vision"
	domainreport "cursor-hackathon/backend/internal/domain/report"
	domainvision "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/config"
)

// ---------------------------------------------------------------------------
// Fake adapters (handler tests build a real CreateReportUseCase)
// ---------------------------------------------------------------------------

// noopAnonymizer passes the image through with fixed dimensions.
type noopAnonymizer struct{}

func (noopAnonymizer) Anonymize(_ context.Context, img []byte) (intake.AnonymizationResult, error) {
	return intake.AnonymizationResult{
		Image:    img,
		Width:    100,
		Height:   100,
		Strategy: domainvision.PIIStrategyAvoidanceByDesign,
	}, nil
}

// okAnalyzer always returns one road_damage detection.
type okAnalyzer struct{}

func (okAnalyzer) Execute(_ context.Context, cmd appvision.AnalyzeCommand) (domainvision.AnalysisResult, error) {
	return domainvision.AnalysisResult{
		AnalysisID: "ana_test",
		ReportID:   cmd.ReportID,
		Detections: []domainvision.Detection{
			{
				NormalizedObjectType: domainvision.TypeRoadDamage,
				Confidence:           0.9,
				Priority:             domainvision.PriorityHigh,
				ReviewStatus:         domainvision.ReviewAutoAccepted,
			},
		},
		KVKKSafe: true,
	}, nil
}

// emptyAnalyzer returns no detections (used when we only want to reach the use
// case without caring about domain result).
type emptyAnalyzer struct{}

func (emptyAnalyzer) Execute(_ context.Context, cmd appvision.AnalyzeCommand) (domainvision.AnalysisResult, error) {
	return domainvision.AnalysisResult{
		AnalysisID: "ana_empty",
		ReportID:   cmd.ReportID,
		KVKKSafe:   true,
	}, nil
}

// handlerMemStore is an in-memory ReportStorePort for handler tests.
type handlerMemStore struct {
	items []domainreport.Report
}

func (s *handlerMemStore) Save(_ context.Context, r domainreport.Report) error {
	for i, existing := range s.items {
		if existing.ReportID == r.ReportID {
			s.items[i] = r
			return nil
		}
	}
	s.items = append(s.items, r)
	return nil
}

func (s *handlerMemStore) Get(_ context.Context, id string) (domainreport.Report, error) {
	for _, r := range s.items {
		if r.ReportID == id {
			return r, nil
		}
	}
	return domainreport.Report{}, domainreport.ErrReportNotFound
}

func (s *handlerMemStore) List(_ context.Context) []domainreport.Report {
	if s.items == nil {
		return []domainreport.Report{}
	}
	return s.items
}

func (s *handlerMemStore) FindRecentDuplicate(_ context.Context, lat, lng float64, pt string) (domainreport.Report, bool) {
	return domainreport.Report{}, false
}

func (s *handlerMemStore) IncrementDuplicate(_ context.Context, id string) (domainreport.Report, error) {
	return domainreport.Report{}, domainreport.ErrReportNotFound
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newTestHandler wires a Handler backed by a real use case with fake adapters.
func newTestHandler(t *testing.T, analyzer intake.AnalyzerPort, store *handlerMemStore) (*Handler, *http.ServeMux) {
	t.Helper()
	rules := config.MustLoad()
	uc := intake.NewCreateReportUseCase(noopAnonymizer{}, analyzer, store, rules)
	h := NewHandler(uc, store)
	mux := http.NewServeMux()
	h.Register(mux)
	return h, mux
}

// buildMultipartRequest builds a multipart POST to path with an optional image
// field. Pass nil for imgBytes to omit the image field entirely.
func buildMultipartRequest(t *testing.T, path string, imgBytes []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if imgBytes != nil {
		part, err := writer.CreateFormFile("image", "test.jpg")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(imgBytes); err != nil {
			t.Fatalf("write image: %v", err)
		}
	}

	writer.Close()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestCreateReport_MissingImage: POST without an image field → 400.
func TestCreateReport_MissingImage(t *testing.T) {
	store := &handlerMemStore{}
	_, mux := newTestHandler(t, okAnalyzer{}, store)

	req := buildMultipartRequest(t, "/api/v1/reports", nil) // no image
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", rr.Code, rr.Body.String())
	}
}

// TestCreateReport_ForbiddenRole: manager cannot create reports → 403.
func TestCreateReport_ForbiddenRole(t *testing.T) {
	store := &handlerMemStore{}
	_, mux := newTestHandler(t, okAnalyzer{}, store)

	req := buildMultipartRequest(t, "/api/v1/reports", []byte{1, 2, 3})
	req.Header.Set("X-Role", "manager")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", rr.Code, rr.Body.String())
	}
}

// TestList_Empty: GET /api/v1/reports on an empty store → 200 with JSON array.
func TestList_Empty(t *testing.T) {
	store := &handlerMemStore{}
	_, mux := newTestHandler(t, okAnalyzer{}, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []domainreport.Report
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// result may be empty or nil — both are valid for an empty store.
	if result == nil {
		result = []domainreport.Report{}
	}
}

// TestGet_NotFound: GET /api/v1/reports/nonexistent → 404.
func TestGet_NotFound(t *testing.T) {
	store := &handlerMemStore{}
	_, mux := newTestHandler(t, okAnalyzer{}, store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/nonexistent", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", rr.Code, rr.Body.String())
	}
}
