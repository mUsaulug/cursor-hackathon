package app

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	report "cursor-hackathon/backend/internal/domain/report"
)

func sampleJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 320, 240))
	for y := 0; y < 240; y++ {
		for x := 0; x < 320; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 90, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return buf.Bytes()
}

func reportForm(t *testing.T, fields map[string]string, img []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	fw, err := mw.CreateFormFile("image", "report.jpg")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	_, _ = fw.Write(img)
	_ = mw.Close()
	return &body, mw.FormDataContentType()
}

// TestE2ECitizenIntake walks the real intake scenario: a resident submits a
// problem photo + note + location; the system anonymizes (KVKK), analyzes,
// classifies, routes to a department, and queues it for operator review.
func TestE2ECitizenIntake(t *testing.T) {
	srv := httptest.NewServer(NewMux())
	defer srv.Close()

	img := sampleJPEG(t)

	t.Run("citizen can file a report", func(t *testing.T) {
		body, ct := reportForm(t, map[string]string{
			"description": "Yolda buyuk cukur var",
			"source_type": report.SourceCitizenMobile,
			"lat":         "41.0082",
			"lng":         "28.9784",
		}, img)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/reports", body)
		req.Header.Set("Content-Type", ct)
		req.Header.Set("X-Role", "citizen")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want 201", resp.StatusCode)
		}
		var rep report.Report
		decode(t, resp, &rep)

		if rep.ReportID == "" || rep.AnalysisID == "" {
			t.Errorf("missing ids: %+v", rep)
		}
		if rep.Status != report.StatusWaitingForReview {
			t.Errorf("status = %q, want waiting_for_review", rep.Status)
		}
		if rep.AssignedDept == "" {
			t.Error("expected a routed department")
		}
		if rep.ProblemType == "" {
			t.Error("expected a problem type")
		}
		if rep.ReporterRole != "citizen" {
			t.Errorf("reporter_role = %q", rep.ReporterRole)
		}
	})

	t.Run("manager role cannot file reports (RBAC)", func(t *testing.T) {
		body, ct := reportForm(t, map[string]string{"description": "x"}, img)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/reports", body)
		req.Header.Set("Content-Type", ct)
		req.Header.Set("X-Role", "manager")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", resp.StatusCode)
		}
	})

	t.Run("reports are listable", func(t *testing.T) {
		var reports []report.Report
		getJSON(t, srv.URL+"/api/v1/reports", &reports)
		if len(reports) < 1 {
			t.Error("expected at least one report")
		}
	})

	t.Run("duplicate near same location is merged", func(t *testing.T) {
		// Two identical-location reports of the same problem -> second merges.
		var first report.Report
		body, ct := reportForm(t, map[string]string{
			"description": "Cukur", "lat": "40.5000", "lng": "29.5000",
		}, img)
		postMultipart(t, srv.URL+"/api/v1/reports", body, ct, "citizen", &first)

		var second report.Report
		body2, ct2 := reportForm(t, map[string]string{
			"description": "Ayni cukur", "lat": "40.5000", "lng": "29.5000",
		}, img)
		postMultipart(t, srv.URL+"/api/v1/reports", body2, ct2, "citizen", &second)

		if second.Status != report.StatusMerged || second.DuplicateOf != first.ReportID {
			t.Errorf("expected second report merged into first; got status=%q dup_of=%q", second.Status, second.DuplicateOf)
		}
	})
}
