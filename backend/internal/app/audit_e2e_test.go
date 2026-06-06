package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	audit "cursor-hackathon/backend/internal/domain/audit"
	report "cursor-hackathon/backend/internal/domain/report"
)

// TestE2EAuditTrail verifies that mutating actions are recorded and that only a
// manager can read the audit log (KVKK accountability).
func TestE2EAuditTrail(t *testing.T) {
	srv := httptest.NewServer(NewMux())
	defer srv.Close()
	img := sampleJPEG(t)

	var rep report.Report
	body, ct := reportForm(t, map[string]string{"description": "Cukur", "lat": "41", "lng": "29"}, img)
	postMultipart(t, srv.URL+"/api/v1/reports", body, ct, "citizen", &rep)

	// Non-manager cannot read the audit log.
	if code := getStatus(t, srv.URL+"/api/v1/audit", "operator"); code != http.StatusForbidden {
		t.Fatalf("operator audit read = %d, want 403", code)
	}

	// Manager reads the log and finds the report POST recorded.
	var entries []audit.Entry
	getJSONRole(t, srv.URL+"/api/v1/audit", "manager", &entries)
	var found bool
	for _, e := range entries {
		if e.Method == http.MethodPost && e.Path == "/api/v1/reports" && e.ActorRole == "citizen" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an audit entry for the citizen report POST; got %d entries", len(entries))
	}
}

func getStatus(t *testing.T, url, role string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("X-Role", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
