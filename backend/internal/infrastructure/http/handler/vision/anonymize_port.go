package vision

import "context"

// AnonymizeOutput is the anonymizer result used before vision inference.
type AnonymizeOutput struct {
	Image          []byte
	Width          int
	Height         int
	RegionsBlurred int
	Strategy       string
	Anonymized     bool
}

// ImageAnonymizer redacts PII before inference on real image bytes.
type ImageAnonymizer interface {
	Anonymize(ctx context.Context, img []byte) (AnonymizeOutput, error)
}

// StreetViewFetcher retrieves a raw Street View frame (must be anonymized before use).
type StreetViewFetcher interface {
	Fetch(ctx context.Context, lat, lng float64) ([]byte, error)
}
