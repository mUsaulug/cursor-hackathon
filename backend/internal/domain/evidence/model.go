// Package evidence is the domain layer for task-completion proof. Field staff
// upload an "after" photo; the system compares it to the "before" analysis to
// decide whether the problem is likely resolved (deterministic diff, not LLM).
package evidence

// CompletionEvidence is the proof + verification for a completed task.
type CompletionEvidence struct {
	EvidenceID       string       `json:"evidence_id"`
	TaskID           string       `json:"task_id"`
	BeforeAnalysisID string       `json:"before_analysis_id,omitempty"`
	AfterAnalysisID  string       `json:"after_analysis_id,omitempty"`
	ImageRef         string       `json:"image_ref,omitempty"` // anonymized derivative only
	AIVerification   Verification `json:"ai_verification"`
	ManagerApproval  string       `json:"manager_approval"` // pending | approved | rejected
	UploadedBy       string       `json:"uploaded_by,omitempty"`
	CreatedAt        string       `json:"created_at"`
}

// Verification is the deterministic before/after outcome.
type Verification string

const (
	VerificationLikelyResolved Verification = "likely_resolved"
	VerificationStillPresent   Verification = "still_present"
	VerificationNeedsHuman     Verification = "needs_human"
)

// Manager approval states.
const (
	ApprovalPending  = "pending"
	ApprovalApproved = "approved"
	ApprovalRejected = "rejected"
)
