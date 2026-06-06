// Package task is the domain layer for municipal work orders created from
// accepted reports. A Task carries assignment, priority, SLA and lifecycle.
package task

// Task is a work order routed to a municipal department.
type Task struct {
	TaskID       string `json:"task_id"`
	ReportID     string `json:"report_id"`
	AssignedDept string `json:"assigned_department"`
	AssignedTo   string `json:"assigned_to,omitempty"`
	Priority     string `json:"priority"`
	Status       Status `json:"status"`
	SLA          string `json:"sla,omitempty"` // e.g. "48h"
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// Status is the task lifecycle state.
type Status string

const (
	StatusCreated          Status = "created"
	StatusAssigned         Status = "assigned"
	StatusStarted          Status = "started"
	StatusEvidenceUploaded Status = "evidence_uploaded"
	StatusAIVerified       Status = "ai_verified"
	StatusCompleted        Status = "completed"
	StatusReopened         Status = "reopened"
)

// CanTransition reports whether moving from one status to another is allowed.
// The state machine keeps the lifecycle honest (no skipping verification).
func CanTransition(from, to Status) bool {
	allowed := map[Status][]Status{
		StatusCreated:          {StatusAssigned},
		StatusAssigned:         {StatusStarted, StatusAssigned},
		StatusStarted:          {StatusEvidenceUploaded},
		StatusEvidenceUploaded: {StatusAIVerified},
		StatusAIVerified:       {StatusCompleted, StatusReopened},
		StatusReopened:         {StatusAssigned},
		StatusCompleted:        {},
	}
	for _, t := range allowed[from] {
		if t == to {
			return true
		}
	}
	return false
}
