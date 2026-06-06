package intake

import (
	"strings"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// textSignalKeywords maps Turkish/English keywords in a citizen description to a
// normalized object type. This is ONLY a supporting signal: it is used when the
// vision model is unsure (unknown/empty). It never overrides a confident model
// decision, keeping the decision chain model-grounded and deterministic.
var textSignalKeywords = map[string][]string{
	domain.TypeRoadDamage:      {"cukur", "çukur", "pothole", "catlak", "çatlak", "asfalt", "yol bozuk", "kaplama"},
	domain.TypeWasteAsset:      {"cop", "çöp", "konteyner", "atik", "atık", "garbage", "trash"},
	domain.TypeSidewalk:        {"kaldirim", "kaldırım", "sidewalk", "yaya yolu", "tretuvar"},
	domain.TypeTrafficSignal:   {"trafik", "lamba", "isik", "ışık", "levha", "tabela", "sinyal", "aydinlatma", "aydınlatma"},
	domain.TypeStreetFurniture: {"bank", "bench", "durak", "park", "oturma"},
}

// classifyText returns a candidate normalized type from free text, or "" if no
// keyword matches.
func classifyText(description string) string {
	d := strings.ToLower(description)
	for objectType, keywords := range textSignalKeywords {
		for _, kw := range keywords {
			if strings.Contains(d, kw) {
				return objectType
			}
		}
	}
	return ""
}
