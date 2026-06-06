// Package identity is the domain layer for users and roles (RBAC). Wave 2 has
// four roles; the MVP uses a simple role token, the product uses masterfabric-go
// JWT/RBAC. Every write is attributable for KVKK accountability (audit trail).
package identity

// Role is a coarse RBAC role.
type Role string

const (
	RoleCitizen    Role = "citizen"
	RoleFieldStaff Role = "field_staff"
	RoleOperator   Role = "operator"
	RoleManager    Role = "manager"
)

// User is an actor in the system.
type User struct {
	UserID string `json:"user_id"`
	Name   string `json:"name,omitempty"`
	Role   Role   `json:"role"`
}

// Capability is a coarse permission checked by use cases/handlers.
type Capability string

const (
	CapCreateReport   Capability = "report.create"
	CapReviewReport   Capability = "report.review"
	CapCreateTask     Capability = "task.create"
	CapUploadEvidence Capability = "task.evidence"
	CapCloseTask      Capability = "task.close"
	CapViewAnalytics  Capability = "analytics.view"
)

// roleCaps maps each role to its capabilities.
var roleCaps = map[Role]map[Capability]bool{
	RoleCitizen: {
		CapCreateReport: true,
	},
	RoleFieldStaff: {
		CapCreateReport:   true,
		CapUploadEvidence: true,
	},
	RoleOperator: {
		CapCreateReport: true,
		CapReviewReport: true,
		CapCreateTask:   true,
	},
	RoleManager: {
		CapCloseTask:     true,
		CapViewAnalytics: true,
	},
}

// Can reports whether a role holds a capability.
func (r Role) Can(c Capability) bool {
	return roleCaps[r][c]
}

// ValidRole reports whether s is a known role.
func ValidRole(s string) bool {
	switch Role(s) {
	case RoleCitizen, RoleFieldStaff, RoleOperator, RoleManager:
		return true
	default:
		return false
	}
}
