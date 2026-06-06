package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	evidence "cursor-hackathon/backend/internal/domain/evidence"
	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
)

// TestE2EFullLifecycle exercises the complete municipal operations loop close to
// real usage: citizen reports -> operator accepts+assigns -> field staff starts
// -> field staff uploads completion evidence (anonymized + verified) -> manager
// closes the task.
func TestE2EFullLifecycle(t *testing.T) {
	srv := httptest.NewServer(NewMux())
	defer srv.Close()
	img := sampleJPEG(t)

	// citizen report
	var rep report.Report
	body, ct := reportForm(t, map[string]string{
		"description": "Cukur var", "lat": "41.1", "lng": "29.1",
	}, img)
	postMultipart(t, srv.URL+"/api/v1/reports", body, ct, "citizen", &rep)

	// operator accept + assign
	var tk task.Task
	postJSONRole(t, srv.URL+"/api/v1/reports/"+rep.ReportID+"/review",
		`{"decision":"accepted","assigned_to":"saha_2"}`, "operator", http.StatusCreated, &tk)

	// field staff start
	var started task.Task
	postJSONRole(t, srv.URL+"/api/v1/tasks/"+tk.TaskID+"/start", ``, "field_staff", http.StatusOK, &started)

	// field staff uploads "after" evidence
	ebody, ect := reportForm(t, map[string]string{"uploaded_by": "saha_2"}, img)
	var ev evidence.CompletionEvidence
	postMultipartTo(t, srv.URL+"/api/v1/tasks/"+tk.TaskID+"/evidence", ebody, ect, "field_staff", &ev)
	if ev.EvidenceID == "" || ev.AfterAnalysisID == "" {
		t.Fatalf("evidence incomplete: %+v", ev)
	}
	if ev.AIVerification == "" {
		t.Error("expected a verification outcome")
	}

	// citizen must NOT be able to close a task (RBAC)
	if code := postJSONStatus(t, srv.URL+"/api/v1/tasks/"+tk.TaskID+"/close",
		`{"decision":"approved"}`, "citizen"); code != http.StatusForbidden {
		t.Fatalf("citizen close status = %d, want 403", code)
	}

	// manager closes -> completed
	var closed task.Task
	postJSONRole(t, srv.URL+"/api/v1/tasks/"+tk.TaskID+"/close",
		`{"decision":"approved"}`, "manager", http.StatusOK, &closed)
	if closed.Status != task.StatusCompleted {
		t.Errorf("status = %q, want completed", closed.Status)
	}

	// manager dashboard reflects the completed work
	var summary struct {
		TotalReports   int `json:"total_reports"`
		TotalTasks     int `json:"total_tasks"`
		CompletedTasks int `json:"completed_tasks"`
	}
	getJSONRole(t, srv.URL+"/api/v1/analytics/summary", "manager", &summary)
	if summary.TotalReports < 1 || summary.CompletedTasks < 1 {
		t.Errorf("analytics summary not reflecting work: %+v", summary)
	}
}

func getJSONRole(t *testing.T, url, role string, out any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("X-Role", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s status = %d", url, resp.StatusCode)
	}
	decode(t, resp, out)
}

// postMultipartTo posts a multipart body expecting 201 Created.
func postMultipartTo(t *testing.T, url string, body interface{ Read([]byte) (int, error) }, contentType, role string, out any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Role", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST %s status = %d, want 201", url, resp.StatusCode)
	}
	decode(t, resp, out)
}
