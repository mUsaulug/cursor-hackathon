// Package vision is the domain layer of the CivicLens Core vision bounded
// context. It holds dependency-free types and rules. No HTTP, no Hugging Face,
// no infrastructure imports may appear here (hexagonal dependency direction).
package vision

// SchemaVersion is the AnalysisResult contract version. Bump on breaking change
// so web/mobile consumers can detect schema drift.
const SchemaVersion = "1.0"

// AnalysisResult is the single schema consumed by the web dashboard and the
// Expo field view. Every adapter output is normalized into this shape.
type AnalysisResult struct {
	SchemaVersion  string        `json:"schema_version"`
	AnalysisID     string        `json:"analysis_id"`
	ReportID       string        `json:"report_id,omitempty"` // Wave 2: links an analysis to a citizen/staff report
	SourceType     string        `json:"source_type"`
	SourceRef      string        `json:"source_ref"`
	Location       *Location     `json:"location,omitempty"`
	ModelID        string        `json:"model_id"`
	ModelMode      string        `json:"model_mode"`
	ImageWidth     int           `json:"image_width"`
	ImageHeight    int           `json:"image_height"`
	RawImageStored bool          `json:"raw_image_stored"`
	Anonymized     bool          `json:"anonymized"`
	KVKKSafe       bool          `json:"kvkk_safe"`
	Privacy        PrivacyReport `json:"privacy"`
	Detections     []Detection   `json:"detections"`
	CreatedAt      string        `json:"created_at"`
	DeletionStatus string        `json:"deletion_status"`
}

// Detection is one normalized detection. Each detection knows which model
// produced it and carries its own UUID for the human review flow.
type Detection struct {
	ID                   string       `json:"id"`
	Label                string       `json:"label"`
	NormalizedObjectType string       `json:"normalized_object_type"`
	Confidence           float64      `json:"confidence"`
	BBox                 BoundingBox  `json:"bbox"`
	ModelID              string       `json:"model_id"`
	ReviewStatus         ReviewStatus `json:"review_status"`
	Priority             Priority     `json:"priority"`
	Reason               string       `json:"reason"`
}

// BoundingBox keeps float coordinates exactly as the model returns them.
// Converting to int loses precision needed for overlay alignment.
type BoundingBox struct {
	XMin float64 `json:"xmin"`
	YMin float64 `json:"ymin"`
	XMax float64 `json:"xmax"`
	YMax float64 `json:"ymax"`
}

// Location is an optional geo coordinate for the analyzed image.
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// PrivacyReport is embedded in AnalysisResult.Privacy. It is the single source
// for KVKK status; there is no separate privacy entity.
type PrivacyReport struct {
	KVKKSafe       bool   `json:"kvkk_safe"`
	RawImageStored bool   `json:"raw_image_stored"`
	Anonymized     bool   `json:"anonymized"`
	DeletionStatus string `json:"deletion_status"`
	BlockedCount   int    `json:"blocked_count"`
	PIIStrategy    string `json:"pii_strategy"`
}

// ReviewRecord is a human override on a single detection.
type ReviewRecord struct {
	DetectionID string `json:"detection_id"`
	ReviewedBy  string `json:"reviewed_by"`
	Decision    string `json:"decision"`
	Note        string `json:"note,omitempty"`
}

// RawDetection is the adapter-agnostic input to the CivicLens Core pipeline.
// Each inference adapter (DETR, precomputed, road-damage) emits these; the
// normalizer turns them into Detection values. Keeping this in the domain lets
// the application layer depend only on domain types.
type RawDetection struct {
	Label      string
	Confidence float64
	BBox       BoundingBox
	ModelID    string
}
