// Package vision (http handler) adapts HTTP requests/responses to the vision
// application use cases. It owns transport concerns only: parsing multipart
// uploads, decoding image dimensions, and JSON encoding. No decision logic.
package vision

import appvision "cursor-hackathon/backend/internal/application/vision"

// errorDTO is the standard JSON error body.
type errorDTO struct {
	Error string `json:"error"`
}

// modelInfoDTO is the transparency payload for GET /model-info (design doc 10).
type modelInfoDTO struct {
	ActiveModes []string        `json:"active_modes"`
	DefaultMode string          `json:"default_mode"`
	Models      []modelEntryDTO `json:"models"`
	Limitations []string        `json:"limitations"`
}

type modelEntryDTO struct {
	ModelID string `json:"model_id"`
	Mode    string `json:"mode"`
	Role    string `json:"role"`
	Live    bool   `json:"live"`
}

// reviewRequestDTO is the body for PATCH /reviews/{detectionId}.
type reviewRequestDTO struct {
	Decision   string `json:"decision"` // "accepted" | "rejected"
	ReviewedBy string `json:"reviewed_by"`
	Note       string `json:"note"`
}

// reviewItemDTO is one needs-review detection in GET /reviews.
type reviewItemDTO struct {
	AnalysisID           string  `json:"analysis_id"`
	DetectionID          string  `json:"detection_id"`
	Label                string  `json:"label"`
	NormalizedObjectType string  `json:"normalized_object_type"`
	Confidence           float64 `json:"confidence"`
	Priority             string  `json:"priority"`
	Reason               string  `json:"reason"`
}

// reportDTO wraps the maintenance report plus the analysis it describes.
type reportDTO struct {
	AnalysisID string                       `json:"analysis_id"`
	Report     *appvision.MaintenanceReport `json:"report"`
}
