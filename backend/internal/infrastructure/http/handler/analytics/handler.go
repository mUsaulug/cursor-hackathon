// Package analytics (http handler) exposes the manager dashboard summary.
package analytics

import (
	"encoding/json"
	"net/http"

	analytics "cursor-hackathon/backend/internal/application/analytics"
	identity "cursor-hackathon/backend/internal/domain/identity"
)

// Handler serves analytics endpoints.
type Handler struct {
	svc *analytics.Service
}

// NewHandler builds the handler.
func NewHandler(svc *analytics.Service) *Handler { return &Handler{svc: svc} }

// Register wires analytics routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/analytics/summary", h.Summary)
}

// Summary returns the manager dashboard aggregation (operator/manager only).
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	role := roleFromRequest(r)
	if !role.Can(identity.CapViewAnalytics) && role != identity.RoleOperator {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "role not permitted"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.svc.Summary(r.Context()))
}

func roleFromRequest(r *http.Request) identity.Role {
	role := r.Header.Get("X-Role")
	if identity.ValidRole(role) {
		return identity.Role(role)
	}
	return identity.RoleCitizen
}
