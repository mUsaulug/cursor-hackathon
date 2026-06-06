// Package report (http handler) exposes Wave 2 intake endpoints. Transport only:
// it parses the multipart report, enforces a coarse role capability, and calls
// the intake use case. KVKK anonymization happens inside the use case.
package report

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	intake "cursor-hackathon/backend/internal/application/intake"
	identity "cursor-hackathon/backend/internal/domain/identity"
	report "cursor-hackathon/backend/internal/domain/report"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

const maxUploadBytes = 12 << 20

// Handler serves report intake and listing.
type Handler struct {
	intake *intake.CreateReportUseCase
	store  intake.ReportStorePort
}

// NewHandler builds the handler.
func NewHandler(uc *intake.CreateReportUseCase, store intake.ReportStorePort) *Handler {
	return &Handler{intake: uc, store: store}
}

// Register wires report routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/reports", h.Create)
	mux.HandleFunc("GET /api/v1/reports", h.List)
	mux.HandleFunc("GET /api/v1/reports/{id}", h.Get)
}

// Create ingests a citizen/staff report (multipart: image + description + location).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	role := roleFromRequest(r)
	if !role.Can(identity.CapCreateReport) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "role cannot create reports"})
		return
	}

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "image file required"})
		return
	}
	defer file.Close()
	img, err := io.ReadAll(io.LimitReader(file, maxUploadBytes))
	if err != nil || len(img) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty image"})
		return
	}

	cmd := intake.CreateReportCommand{
		Image:        img,
		Description:  r.FormValue("description"),
		SourceType:   defaultStr(r.FormValue("source_type"), report.SourceCitizenMobile),
		ReporterRole: string(role),
		ReporterID:   r.FormValue("reporter_id"),
		Location:     parseLocation(r.FormValue("lat"), r.FormValue("lng")),
	}

	rep, err := h.intake.Execute(r.Context(), cmd)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rep)
}

// List returns all reports (newest-first not enforced; in insertion order).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.List(r.Context()))
}

// Get returns one report by id.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	rep, err := h.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

func roleFromRequest(r *http.Request) identity.Role {
	role := r.Header.Get("X-Role")
	if identity.ValidRole(role) {
		return identity.Role(role)
	}
	return identity.RoleCitizen // default: anonymous citizen reporter
}

func parseLocation(latStr, lngStr string) *report.Location {
	if latStr == "" || lngStr == "" {
		return nil
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		return nil
	}
	return &report.Location{Lat: lat, Lng: lng}
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, report.ErrReportNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrNoImage), errors.Is(err, domain.ErrUnsupportedImage):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}
}
