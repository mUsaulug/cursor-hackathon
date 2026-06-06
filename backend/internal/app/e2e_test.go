package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// TestE2EVisionFlow drives the fully wired API like a client would and asserts
// the design-doc §9.8 evaluation checkpoints end to end.
func TestE2EVisionFlow(t *testing.T) {
	srv := httptest.NewServer(NewMux())
	defer srv.Close()

	t.Run("health", func(t *testing.T) {
		var body map[string]string
		getJSON(t, srv.URL+"/health", &body)
		if body["status"] != "ok" {
			t.Fatalf("health = %v", body)
		}
	})

	t.Run("analyze traffic sample filters PII and assigns priority", func(t *testing.T) {
		var res domain.AnalysisResult
		postJSON(t, srv.URL+"/api/v1/vision/analyze?source_ref=sample_street_traffic", &res)

		if res.SchemaVersion != domain.SchemaVersion {
			t.Errorf("schema_version = %q", res.SchemaVersion)
		}
		if res.RawImageStored {
			t.Error("raw_image_stored must be false")
		}
		if !res.KVKKSafe {
			t.Error("kvkk_safe must be true")
		}
		// person is in the traffic fixture and must be blocked (counted).
		if res.Privacy.BlockedCount < 1 {
			t.Errorf("blocked_count = %d, want >=1 (person filtered)", res.Privacy.BlockedCount)
		}
		// No blocked/hidden classes may leak into detections.
		for _, d := range res.Detections {
			if d.Label == "person" || d.Label == "car" || d.Label == "bicycle" {
				t.Errorf("leaked PII/hidden label %q", d.Label)
			}
			if d.Priority == "" || string(d.Priority) == "needs_review" {
				t.Errorf("invalid priority %q for %q", d.Priority, d.NormalizedObjectType)
			}
			if d.ReviewStatus == "" {
				t.Errorf("missing review_status for %q", d.NormalizedObjectType)
			}
		}
	})

	t.Run("road damage hero: high priority + low-confidence needs_review", func(t *testing.T) {
		var res domain.AnalysisResult
		postJSON(t, srv.URL+"/api/v1/vision/analyze?source_ref=sample_road_damage&mode=road_damage", &res)

		if res.ModelMode != domain.ModelModeRoadDamage {
			t.Errorf("model_mode = %q, want road_damage", res.ModelMode)
		}
		var sawHigh, sawNeedsReview bool
		for _, d := range res.Detections {
			if d.NormalizedObjectType != domain.TypeRoadDamage {
				t.Errorf("unexpected type %q", d.NormalizedObjectType)
			}
			if d.Priority == domain.PriorityHigh {
				sawHigh = true
			}
			if d.ReviewStatus == domain.ReviewNeedsReview {
				sawNeedsReview = true
			}
		}
		if !sawHigh {
			t.Error("expected at least one high-priority road damage detection")
		}
		if !sawNeedsReview {
			t.Error("expected a low-confidence detection routed to needs_review")
		}
	})

	t.Run("demo-results returns two reproducible analyses", func(t *testing.T) {
		var out []domain.AnalysisResult
		getJSON(t, srv.URL+"/api/v1/vision/demo-results", &out)
		if len(out) != 2 {
			t.Fatalf("demo-results len = %d, want 2", len(out))
		}
	})

	t.Run("reviews list then override flips status", func(t *testing.T) {
		var items []struct {
			DetectionID string `json:"detection_id"`
		}
		getJSON(t, srv.URL+"/api/v1/vision/reviews", &items)
		if len(items) == 0 {
			t.Fatal("expected needs_review items from prior analyses")
		}
		id := items[0].DetectionID

		req, _ := http.NewRequest(http.MethodPatch,
			srv.URL+"/api/v1/vision/reviews/"+id,
			jsonBody(`{"decision":"accepted","reviewed_by":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PATCH: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("PATCH status = %d", resp.StatusCode)
		}
		var updated domain.AnalysisResult
		_ = json.NewDecoder(resp.Body).Decode(&updated)
		var found bool
		for _, d := range updated.Detections {
			if d.ID == id {
				found = true
				if d.ReviewStatus != domain.ReviewAutoAccepted {
					t.Errorf("status after accept = %q", d.ReviewStatus)
				}
			}
		}
		if !found {
			t.Error("reviewed detection not found in updated analysis")
		}
	})

	t.Run("summary report degrades gracefully", func(t *testing.T) {
		var rep struct {
			AnalysisID string `json:"analysis_id"`
			Report     *struct {
				RiskLevel string `json:"risk_level"`
			} `json:"report"`
		}
		getJSON(t, srv.URL+"/api/v1/vision/summary", &rep)
		if rep.Report == nil || rep.Report.RiskLevel == "" {
			t.Errorf("expected a report with risk level, got %+v", rep)
		}
	})
}
