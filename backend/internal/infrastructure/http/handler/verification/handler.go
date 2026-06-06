// Package verification (http handler) exposes completion-evidence upload (field
// staff) and task close/reopen (manager).
package verification

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	verification "cursor-hackathon/backend/internal/application/verification"
	identity "cursor-hackathon/backend/internal/domain/identity"
	report "cursor-hackathon/backend/internal/domain/report"
	task "cursor-hackathon/backend/internal/domain/task"
)

const maxUploadBytes = 12 << 20

// Handler serves verification endpoints.
type Handler struct {
	svc *verification.Service
}

// NewHandler builds the handler.
func NewHandler(svc *verification.Service) *Handler { return &Handler{svc: svc} }

// Register wires evidence + close routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tasks/{id}/evidence", h.UploadEvidence)
	mux.HandleFunc("POST /api/v1/tasks/{id}/close", h.Close)
}

// UploadEvidence accepts an "after" photo from field staff (multipart image).
func (h *Handler) UploadEvidence(w http.ResponseWriter, r *http.Request) {
	if !roleFromRequest(r).Can(identity.CapUploadEvidence) {
		forbid(w)
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
	ev, err := h.svc.UploadEvidence(r.Context(), r.PathValue("id"), img, r.FormValue("uploaded_by"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ev)
}

// Close lets a manager complete or reopen a verified task.
func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	if !roleFromRequest(r).Can(identity.CapCloseTask) {
		forbid(w)
		return
	}
	var body struct {
		Decision string `json:"decision"` // approved/completed | rejected/reopened
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	t, err := h.svc.CloseTask(r.Context(), r.PathValue("id"), body.Decision)
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
	case errors.Is(err, task.ErrTaskNotFound), errors.Is(err, report.ErrReportNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, verification.ErrInvalidState):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}
}
