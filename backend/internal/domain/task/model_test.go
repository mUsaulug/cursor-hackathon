package task

import "testing"

func TestCanTransition(t *testing.T) {
	valid := [][2]Status{
		{StatusCreated, StatusAssigned},
		{StatusAssigned, StatusStarted},
		{StatusStarted, StatusEvidenceUploaded},
		{StatusEvidenceUploaded, StatusAIVerified},
		{StatusAIVerified, StatusCompleted},
		{StatusAIVerified, StatusReopened},
		{StatusReopened, StatusAssigned},
	}
	for _, c := range valid {
		if !CanTransition(c[0], c[1]) {
			t.Errorf("expected %s -> %s to be allowed", c[0], c[1])
		}
	}

	invalid := [][2]Status{
		{StatusCreated, StatusCompleted},   // cannot skip verification
		{StatusStarted, StatusCompleted},   // must upload evidence first
		{StatusCompleted, StatusAssigned},  // terminal
		{StatusAssigned, StatusAIVerified}, // must go through evidence
	}
	for _, c := range invalid {
		if CanTransition(c[0], c[1]) {
			t.Errorf("expected %s -> %s to be rejected", c[0], c[1])
		}
	}
}
