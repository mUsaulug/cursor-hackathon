package vision

import (
	"fmt"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// buildReason produces a short, deterministic Turkish rationale for a detection.
// It is template-based (no LLM): the decision chain is the source of truth and
// the optional OpenRouter reasoner only rewrites prose later (design doc 8).
func buildReason(objectType string, priority domain.Priority, status domain.ReviewStatus, mapped bool) string {
	if !mapped || objectType == domain.TypeUnknown {
		return "Tanimlanamayan nesne; saha personeli tarafindan dogrulanmali."
	}

	label := objectTypeLabelTR(objectType)
	switch status {
	case domain.ReviewNeedsReview:
		return fmt.Sprintf("%s tespit edildi ancak guven esigi altinda; insan onayi gerekiyor.", label)
	default: // auto_accepted
		switch priority {
		case domain.PriorityCritical:
			return fmt.Sprintf("%s tespit edildi; kritik bakim onceligi.", label)
		case domain.PriorityHigh:
			return fmt.Sprintf("%s tespit edildi; yuksek bakim onceligi.", label)
		case domain.PriorityMedium:
			return fmt.Sprintf("%s yuksek guvenle tespit edildi; orta oncelik.", label)
		default:
			return fmt.Sprintf("%s tespit edildi; dusuk oncelikli envanter kaydi.", label)
		}
	}
}

func objectTypeLabelTR(objectType string) string {
	switch objectType {
	case domain.TypeTrafficSignal:
		return "Trafik altyapisi"
	case domain.TypeRoadDamage:
		return "Yol hasari"
	case domain.TypeSidewalk:
		return "Kaldirim"
	case domain.TypeStreetFurniture:
		return "Sehir mobilyasi"
	case domain.TypeWasteAsset:
		return "Atik altyapisi"
	default:
		return "Kentsel nesne"
	}
}

// buildAnalysisResult assembles the final AnalysisResult from a processed
// pipeline output and request metadata. raw_image_stored is always false and
// kvkk_safe mirrors the privacy report (design doc 5.7, 6).
func buildAnalysisResult(
	analysisID string,
	cmd AnalyzeCommand,
	modelID string,
	modelMode string,
	processed ProcessResult,
	createdAt string,
) domain.AnalysisResult {
	return domain.AnalysisResult{
		SchemaVersion:  domain.SchemaVersion,
		AnalysisID:     analysisID,
		SourceType:     cmd.SourceType,
		SourceRef:      cmd.SourceRef,
		Location:       cmd.Location,
		ModelID:        modelID,
		ModelMode:      modelMode,
		ImageWidth:     cmd.ImageWidth,
		ImageHeight:    cmd.ImageHeight,
		RawImageStored: false,
		Anonymized:     processed.Privacy.Anonymized,
		KVKKSafe:       processed.Privacy.KVKKSafe,
		Privacy:        processed.Privacy,
		Detections:     processed.Detections,
		CreatedAt:      createdAt,
		DeletionStatus: domain.DeletionRawNotPersisted,
	}
}
