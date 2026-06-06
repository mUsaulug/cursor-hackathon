// Package report is the domain layer for citizen/staff problem reports — the
// real data source of CivicLens Wave 2. A Report is the unit of intake; it links
// to vision AnalysisResult(s) and, once accepted, to a Task. Dependency-free.
package report

// Report is a single urban-problem report from a citizen or field staff member.
type Report struct {
	ReportID       string    `json:"report_id"`
	SourceType     string    `json:"source_type"`   // citizen_mobile | staff_mobile | web
	ReporterRole   string    `json:"reporter_role"` // citizen | field_staff
	ReporterID     string    `json:"reporter_id,omitempty"`
	Description    string    `json:"description"`
	Location       *Location `json:"location,omitempty"`
	AddressContext string    `json:"address_context,omitempty"`
	ImageRef       string    `json:"image_ref,omitempty"` // anonymized derivative only
	AnalysisID     string    `json:"analysis_id,omitempty"`
	ProblemType    string    `json:"problem_type,omitempty"`
	Priority       string    `json:"priority,omitempty"`
	ReviewStatus   string    `json:"review_status,omitempty"`
	AssignedDept   string    `json:"assigned_department,omitempty"`
	DuplicateOf    string    `json:"duplicate_of,omitempty"`
	DuplicateCount int       `json:"duplicate_count"` // number of merged duplicate reports
	Status         Status    `json:"status"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at,omitempty"`
}

// Location is a geo coordinate (kept local to avoid cross-context coupling).
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Status is the report lifecycle state.
type Status string

const (
	StatusCreated          Status = "created"
	StatusAnonymized       Status = "anonymized"
	StatusAIAnalyzed       Status = "ai_analyzed"
	StatusDedupChecked     Status = "dedup_checked"
	StatusWaitingForReview Status = "waiting_for_review"
	StatusTaskCreated      Status = "task_created"
	StatusRejected         Status = "rejected"
	StatusMerged           Status = "merged"
)

// Source types and reporter roles.
const (
	SourceCitizenMobile = "citizen_mobile"
	SourceStaffMobile   = "staff_mobile"
	SourceWeb           = "web"

	ReporterCitizen    = "citizen"
	ReporterFieldStaff = "field_staff"
)
