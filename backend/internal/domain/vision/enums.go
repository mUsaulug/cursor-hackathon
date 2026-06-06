package vision

// Priority answers "how urgent is maintenance?". It is a SEPARATE concept from
// ReviewStatus. Priority is never "needs_review" (design doc 5.5).
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// ReviewStatus answers "does a human need to confirm this?" (design doc 5.6).
type ReviewStatus string

const (
	ReviewAutoAccepted ReviewStatus = "auto_accepted"
	ReviewNeedsReview  ReviewStatus = "needs_review"
	ReviewRejected     ReviewStatus = "rejected"
)

// ModelMode records which inference path produced the result, for transparency
// in the dashboard (design doc 6, 7.3, 12).
const (
	ModelModeLiveHF       = "live_hf"
	ModelModeRoadDamage   = "road_damage"
	ModelModeSegmentation = "segmentation"
	ModelModeOpenVocab    = "open_vocabulary"
	ModelModePrecomputed  = "precomputed"
)

// SourceType records where the analyzed image came from.
const (
	SourceTypeUpload     = "upload"
	SourceTypeSample     = "sample"
	SourceTypeStreetView = "streetview"
)

// PIIStrategy is surfaced to the jury in PrivacyReport (design doc 5.3).
const (
	PIIStrategyAvoidanceByDesign = "avoidance_by_design"
	PIIStrategyBlurApplied       = "blur_applied"
)

// Normalized urban object types (design doc 5.2). Mirrors ontology.yaml.
const (
	TypeTrafficSignal   = "traffic_signal"
	TypeRoadDamage      = "road_damage"
	TypeSidewalk        = "sidewalk"
	TypeStreetFurniture = "street_furniture"
	TypeWasteAsset      = "waste_asset"
	TypeUnknown         = "unknown"
)

// DeletionStatus values for the privacy report.
const (
	DeletionRawNotPersisted = "raw_image_not_persisted"
)

// ReviewDecision values for ReviewRecord.Decision.
const (
	ReviewDecisionAccepted = "accepted"
	ReviewDecisionRejected = "rejected"
)
