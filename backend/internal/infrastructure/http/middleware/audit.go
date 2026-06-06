// Package middleware provides HTTP middleware. The audit middleware records
// every mutating request (POST/PUT/PATCH/DELETE) with the acting role, target,
// and resulting status for KVKK accountability.
package middleware

import (
	"context"
	"net/http"
	"time"

	audit "cursor-hackathon/backend/internal/domain/audit"
	"cursor-hackathon/backend/internal/shared/idgen"
)

// Recorder persists audit entries.
type Recorder interface {
	Record(ctx context.Context, e audit.Entry)
}

// statusWriter captures the response status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Audit wraps a handler and records mutating requests.
func Audit(rec Recorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isMutating(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)

			role := r.Header.Get("X-Role")
			if role == "" {
				role = "anonymous"
			}
			rec.Record(r.Context(), audit.Entry{
				ID:        idgen.NewUUID(),
				ActorRole: role,
				Method:    r.Method,
				Path:      r.URL.Path,
				Status:    sw.status,
				At:        time.Now().Format(time.RFC3339),
			})
		})
	}
}

func isMutating(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
