package vision

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

// Handler serves the vision API. It depends only on application ports.
type Handler struct {
	analyze    *appvision.AnalyzeImageUseCase
	store      appvision.AnalysisStorePort
	reasoner   appvision.ReasonerPort
	anonymizer ImageAnonymizer
	streetView StreetViewFetcher
	models     []modelEntryDTO
	modes      []string
}

// Deps bundles handler dependencies for wiring.
type Deps struct {
	Analyze    *appvision.AnalyzeImageUseCase
	Store      appvision.AnalysisStorePort
	Reasoner   appvision.ReasonerPort
	Anonymizer ImageAnonymizer
	StreetView StreetViewFetcher
	Models     []ModelDescriptor
	Modes      []string
}

// ModelDescriptor is static metadata surfaced via /model-info.
type ModelDescriptor struct {
	ModelID string
	Mode    string
	Role    string
	Live    bool
}

// NewHandler builds the handler.
func NewHandler(d Deps) *Handler {
	models := make([]modelEntryDTO, 0, len(d.Models))
	for _, m := range d.Models {
		models = append(models, modelEntryDTO{ModelID: m.ModelID, Mode: m.Mode, Role: m.Role, Live: m.Live})
	}
	return &Handler{
		analyze:    d.Analyze,
		store:      d.Store,
		reasoner:   d.Reasoner,
		anonymizer: d.Anonymizer,
		streetView: d.StreetView,
		models:     models,
		modes:      d.Modes,
	}
}

// Register wires the vision routes onto a Go 1.22 method-aware ServeMux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/vision/analyze", h.Analyze)
	mux.HandleFunc("GET /api/v1/vision/analyze/{id}", h.GetAnalysis)
	mux.HandleFunc("GET /api/v1/vision/demo-results", h.DemoResults)
	mux.HandleFunc("GET /api/v1/vision/model-info", h.ModelInfo)
	mux.HandleFunc("GET /api/v1/vision/privacy-report", h.PrivacyReport)
	mux.HandleFunc("GET /api/v1/vision/summary", h.Summary)
	mux.HandleFunc("GET /api/v1/vision/reviews", h.ListReviews)
	mux.HandleFunc("PATCH /api/v1/vision/reviews/{detectionId}", h.Review)
	mux.HandleFunc("POST /api/v1/vision/report", h.Report)
}

// Analyze runs an analysis from an uploaded image or a sample request.
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	cmd, err := parseAnalyzeRequest(r)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if err := h.prepareImage(r.Context(), r, &cmd); err != nil {
		writeDomainError(w, err)
		return
	}
	result, err := h.analyze.Execute(r.Context(), cmd)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetAnalysis returns a stored analysis by id (polling path).
func (h *Handler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	result, err := h.store.Get(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// DemoResults runs the two precomputed scenes (traffic + road damage) so the
// dashboard always has a reliable, reproducible demo payload.
func (h *Handler) DemoResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	refs := []string{"sample_street_traffic", "sample_road_damage"}
	out := make([]domain.AnalysisResult, 0, len(refs))
	for _, ref := range refs {
		res, err := h.analyze.Execute(ctx, appvision.AnalyzeCommand{
			SourceType:  domain.SourceTypeSample,
			SourceRef:   ref,
			ModelMode:   domain.ModelModePrecomputed,
			ImageWidth:  sampleWidth,
			ImageHeight: sampleHeight,
		})
		if err != nil {
			writeDomainError(w, err)
			return
		}
		out = append(out, res)
	}
	writeJSON(w, http.StatusOK, out)
}

// ModelInfo reports which models/modes are active and their limitations.
func (h *Handler) ModelInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, modelInfoDTO{
		ActiveModes: h.modes,
		DefaultMode: domain.ModelModePrecomputed,
		Models:      h.models,
		Limitations: []string{
			"DETR (COCO) cannot detect road damage; pothole detection is precomputed RDD2022 inference.",
			"Live HF inference requires HF_API_TOKEN; absent token falls back to precomputed.",
			"KVKK: person/bicycle/motorcycle are blocked; car/truck/bus hidden by default.",
		},
	})
}

// PrivacyReport returns the latest analysis privacy report, or an empty-safe one.
func (h *Handler) PrivacyReport(w http.ResponseWriter, r *http.Request) {
	latest, ok := h.store.Latest(r.Context())
	if !ok {
		writeJSON(w, http.StatusOK, domain.PrivacyReport{
			KVKKSafe:       true,
			RawImageStored: false,
			DeletionStatus: domain.DeletionRawNotPersisted,
			PIIStrategy:    domain.PIIStrategyAvoidanceByDesign,
		})
		return
	}
	writeJSON(w, http.StatusOK, latest.Privacy)
}

// Summary returns a maintenance report for the latest analysis.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	latest, ok := h.store.Latest(r.Context())
	if !ok {
		writeJSON(w, http.StatusOK, reportDTO{})
		return
	}
	report := h.buildReport(r.Context(), latest)
	writeJSON(w, http.StatusOK, reportDTO{AnalysisID: latest.AnalysisID, Report: report})
}

// ListReviews returns all needs-review detections across stored analyses.
func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	items := make([]reviewItemDTO, 0)
	for _, a := range h.store.List(r.Context()) {
		for _, d := range a.Detections {
			if d.ReviewStatus == domain.ReviewNeedsReview {
				items = append(items, reviewItemDTO{
					AnalysisID:           a.AnalysisID,
					DetectionID:          d.ID,
					Label:                d.Label,
					NormalizedObjectType: d.NormalizedObjectType,
					Confidence:           d.Confidence,
					Priority:             string(d.Priority),
					Reason:               d.Reason,
				})
			}
		}
	}
	writeJSON(w, http.StatusOK, items)
}

// Review applies a human decision to a detection.
func (h *Handler) Review(w http.ResponseWriter, r *http.Request) {
	detectionID := r.PathValue("detectionId")
	var req reviewRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: "invalid review body"})
		return
	}
	if req.Decision != domain.ReviewDecisionAccepted && req.Decision != domain.ReviewDecisionRejected {
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: "decision must be 'accepted' or 'rejected'"})
		return
	}
	updated, err := h.store.ApplyReview(r.Context(), domain.ReviewRecord{
		DetectionID: detectionID,
		ReviewedBy:  req.ReviewedBy,
		Decision:    req.Decision,
		Note:        req.Note,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Report generates a maintenance report for a given or latest analysis.
func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AnalysisID string `json:"analysis_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	var (
		result domain.AnalysisResult
		ok     bool
	)
	if body.AnalysisID != "" {
		res, err := h.store.Get(r.Context(), body.AnalysisID)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		result, ok = res, true
	} else {
		result, ok = h.store.Latest(r.Context())
	}
	if !ok {
		writeJSON(w, http.StatusOK, reportDTO{})
		return
	}
	report := h.buildReport(r.Context(), result)
	writeJSON(w, http.StatusOK, reportDTO{AnalysisID: result.AnalysisID, Report: report})
}

// buildReport runs the reasoner if configured; it never fails the request.
func (h *Handler) buildReport(ctx context.Context, result domain.AnalysisResult) *appvision.MaintenanceReport {
	if h.reasoner == nil {
		return nil
	}
	report, err := h.reasoner.GenerateMaintenanceReport(ctx, result)
	if err != nil {
		return nil
	}
	return report
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNoImage), errors.Is(err, domain.ErrUnsupportedImage):
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: err.Error()})
	case errors.Is(err, domain.ErrAnalysisNotFound), errors.Is(err, domain.ErrDetectionNotFound):
		writeJSON(w, http.StatusNotFound, errorDTO{Error: err.Error()})
	case errors.Is(err, domain.ErrUnknownModelMode):
		writeJSON(w, http.StatusUnprocessableEntity, errorDTO{Error: err.Error()})
	case errors.Is(err, errStreetViewUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, errorDTO{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorDTO{Error: err.Error()})
	}
}
