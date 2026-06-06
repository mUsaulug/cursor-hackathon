package vision

import "errors"

// Domain errors. Infrastructure and application layers wrap or translate these
// into HTTP status codes; the domain itself stays transport-agnostic.
var (
	// ErrNoImage is returned when an analyze request carries no image bytes.
	ErrNoImage = errors.New("vision: no image provided")

	// ErrUnsupportedImage is returned when image bytes cannot be decoded.
	ErrUnsupportedImage = errors.New("vision: unsupported or corrupt image")

	// ErrUnknownModelMode is returned when the requested model mode has no
	// registered adapter in the model router.
	ErrUnknownModelMode = errors.New("vision: unknown model mode")

	// ErrInferenceUnavailable is returned when no inference path (live or
	// precomputed) can produce a result.
	ErrInferenceUnavailable = errors.New("vision: inference unavailable")

	// ErrAnalysisNotFound is returned by the store when an analysis ID is unknown.
	ErrAnalysisNotFound = errors.New("vision: analysis not found")

	// ErrDetectionNotFound is returned when a review targets an unknown detection.
	ErrDetectionNotFound = errors.New("vision: detection not found")
)
