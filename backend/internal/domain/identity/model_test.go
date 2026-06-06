package identity

import "testing"

func TestRoleCapabilities(t *testing.T) {
	if !RoleCitizen.Can(CapCreateReport) {
		t.Error("citizen should create reports")
	}
	if RoleCitizen.Can(CapCreateTask) {
		t.Error("citizen must not create tasks")
	}
	if !RoleOperator.Can(CapReviewReport) || !RoleOperator.Can(CapCreateTask) {
		t.Error("operator should review and create tasks")
	}
	if !RoleFieldStaff.Can(CapUploadEvidence) {
		t.Error("field staff should upload evidence")
	}
	if RoleFieldStaff.Can(CapReviewReport) {
		t.Error("field staff must not review reports")
	}
	if !RoleManager.Can(CapCloseTask) || !RoleManager.Can(CapViewAnalytics) {
		t.Error("manager should close tasks and view analytics")
	}
}

func TestValidRole(t *testing.T) {
	if !ValidRole("operator") {
		t.Error("operator should be valid")
	}
	if ValidRole("admin") {
		t.Error("admin is not a known role")
	}
}
