package verification

import (
	"testing"

	evidence "cursor-hackathon/backend/internal/domain/evidence"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

func det(objectType string, conf float64) domain.Detection {
	return domain.Detection{NormalizedObjectType: objectType, Confidence: conf, ReviewStatus: domain.ReviewAutoAccepted}
}

func TestVerify(t *testing.T) {
	cases := []struct {
		name       string
		beforeType string
		after      []domain.Detection
		want       evidence.Verification
	}{
		{"resolved when before type absent", domain.TypeRoadDamage,
			[]domain.Detection{det(domain.TypeTrafficSignal, 0.9)}, evidence.VerificationLikelyResolved},
		{"still present when same type confident", domain.TypeRoadDamage,
			[]domain.Detection{det(domain.TypeRoadDamage, 0.8)}, evidence.VerificationStillPresent},
		{"resolved when same type but low confidence", domain.TypeRoadDamage,
			[]domain.Detection{det(domain.TypeRoadDamage, 0.2)}, evidence.VerificationLikelyResolved},
		{"needs human when before unknown", domain.TypeUnknown,
			[]domain.Detection{det(domain.TypeRoadDamage, 0.9)}, evidence.VerificationNeedsHuman},
		{"resolved when after empty", domain.TypeRoadDamage,
			nil, evidence.VerificationLikelyResolved},
	}
	for _, c := range cases {
		if got := verify(c.beforeType, c.after); got != c.want {
			t.Errorf("%s: verify = %q, want %q", c.name, got, c.want)
		}
	}
}
