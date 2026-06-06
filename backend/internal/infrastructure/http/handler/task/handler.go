// Package task (http handler) exposes operator/field-staff task endpoints and
// the operator report-review action that turns a report into a work order.
package task

import (
	"encoding/json"
	"errors"
	"net/http"

	tasking "cursor-hackathon/backend/internal/application/tasking"
	identity "cursor-hackathon/backend/internal/domain/identity"
	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
)

// Handler serves tasking endpoints.
type Handler struct {
	svc *tasking.Service
}

// NewHandler builds the handler.
func NewHandler(svc *tasking.Service) *Handler { return &Handler{svc: svc} }

// Register wires task + report-review routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/reports/{id}/review", h.Review) // operator accept/reject -> task
	mux.HandleFunc("POST /api/v1/tasks", h.Create)               // operator: report -> task
	mux.HandleFunc("GET /api/v1/tasks", h.List)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.Get)
	mux.HandleFunc("POST /api/v1/tasks/{id}/assign", h.Assign)
	mux.HandleFunc("POST /api/v1/tasks/{id}/start", h.Start)
}

// Review is the operator decision on a report: accept (-> task) or reject.
func (h *Handler) Review(w http.ResponseWriter, r *http.Request) {
	if !roleFromRequest(r).Can(identity.CapReviewReport) {
		forbid(w)
		return
	}
	var body struct {
		Decision   string `json:"decision"` // accepted | rejected
		AssignedTo string `json:"assigned_to"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	id := r.PathValue("id")

	switch body.Decision {
	case "accepted":
		t, err := h.svc.CreateFromReport(r.Context(), id, body.AssignedTo)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, t)
	case "rejected":
		rep, err := h.svc.RejectReport(r.Context(), id)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, rep)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "decision must be 'accepted' or 'rejected'"})
	}
}

// Create turns a report into a task (operator).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if !roleFromRequest(r).Can(identity.CapCreateTask) {
		forbid(w)
		return
	}
	var body struct {
		ReportID   string `json:"report_id"`
		AssignedTo string `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ReportID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "report_id required"})
		return
	}
	t, err := h.svc.CreateFromReport(r.Context(), body.ReportID, body.AssignedTo)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

// List returns tasks, optionally filtered by ?assigned_to=.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.List(r.Context(), r.URL.Query().Get("assigned_to")))
}

// Get returns a task by id.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// Assign assigns a task to a team/person (operator).
func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	if !roleFromRequest(r).Can(identity.CapCreateTask) {
		forbid(w)
		return
	}
	var body struct {
		AssignedTo string `json:"assigned_to"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	t, err := h.svc.Assign(r.Context(), r.PathValue("id"), body.AssignedTo)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// Start marks a task started (field staff).
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	if !roleFromRequest(r).Can(identity.CapUploadEvidence) {
		forbid(w)
		return
	}
	t, err := h.svc.Start(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func roleFromRequest(r *http.Request) identity.Role {
	role := r.Header.Get("X-Role")
	if identity.ValidRole(role) {
		return identity.Role(role)
	}
	return identity.RoleCitizen
}

func forbid(w http.ResponseWriter) {
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "role not permitted"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, report.ErrReportNotFound), errors.Is(err, task.ErrTaskNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, tasking.ErrInvalidTransition):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}
}
