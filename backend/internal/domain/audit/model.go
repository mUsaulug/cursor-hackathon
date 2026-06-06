// Package audit records who did what, when — required for KVKK accountability
// and municipal traceability (every mutating action is attributable).
package audit

// Entry is a single audit record for a mutating request.
type Entry struct {
	ID        string `json:"id"`
	ActorRole string `json:"actor_role"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	At        string `json:"at"`
}
