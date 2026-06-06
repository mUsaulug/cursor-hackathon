package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
)

// TestE2ETaskingFlow walks the municipal operations scenario across roles:
// citizen files a report -> operator accepts and assigns -> field staff starts.
func TestE2ETaskingFlow(t *testing.T) {
	srv := httptest.NewServer(NewMux())
	defer srv.Close()
	img := sampleJPEG(t)

	// 1. Citizen files a report.
	var rep report.Report
	body, ct := reportForm(t, map[string]string{
		"description": "Kaldirim kirik", "lat": "39.9", "lng": "32.8",
	}, img)
	postMultipart(t, srv.URL+"/api/v1/reports", body, ct, "citizen", &rep)

	// 2. A field_staff role must NOT be able to review (RBAC).
	if code := postJSONStatus(t, srv.URL+"/api/v1/reports/"+rep.ReportID+"/review",
		`{"decision":"accepted"}`, "field_staff"); code != http.StatusForbidden {
		t.Fatalf("field_staff review status = %d, want 403", code)
	}

	// 3. Operator accepts and assigns -> task created (assigned).
	var tk task.Task
	postJSONRole(t, srv.URL+"/api/v1/reports/"+rep.ReportID+"/review",
		`{"decision":"accepted","assigned_to":"saha_ekip_1"}`, "operator", http.StatusCreated, &tk)
	if tk.TaskID == "" || tk.Status != task.StatusAssigned {
		t.Fatalf("task not assigned: %+v", tk)
	}
	if tk.AssignedDept == "" || tk.SLA == "" {
		t.Errorf("task missing dept/sla: %+v", tk)
	}

	// 4. Field staff starts the task (assigned -> started).
	var started task.Task
	postJSONRole(t, srv.URL+"/api/v1/tasks/"+tk.TaskID+"/start", ``, "field_staff", http.StatusOK, &started)
	if started.Status != task.StatusStarted {
		t.Errorf("status = %q, want started", started.Status)
	}

	// 5. Field staff sees the task in their queue.
	var queue []task.Task
	getJSON(t, srv.URL+"/api/v1/tasks?assigned_to=saha_ekip_1", &queue)
	if len(queue) != 1 {
		t.Errorf("queue len = %d, want 1", len(queue))
	}

	// 6. Invalid transition is rejected (cannot start an already-started task).
	if code := postJSONStatus(t, srv.URL+"/api/v1/tasks/"+tk.TaskID+"/start", ``, "field_staff"); code != http.StatusConflict {
		t.Errorf("re-start status = %d, want 409", code)
	}
}

func postJSONStatus(t *testing.T, url, body, role string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func postJSONRole(t *testing.T, url, body, role string, wantStatus int, out any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Fatalf("POST %s status = %d, want %d", url, resp.StatusCode, wantStatus)
	}
	if out != nil {
		decode(t, resp, out)
	}
}
