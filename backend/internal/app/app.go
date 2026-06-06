// Package app wires the CivicLens vision bounded context: it loads rules, builds
// the pipeline, registers inference adapters into the model router, and exposes
// an http.ServeMux with the vision routes plus health endpoints. main.go keeps
// only process concerns (CORS, listen). This mirrors how the masterfabric-go
// composition root would assemble a bounded context.
package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	intake "cursor-hackathon/backend/internal/application/intake"
	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/infrastructure/anonymizer"
	"cursor-hackathon/backend/internal/infrastructure/demo"
	reporthttp "cursor-hackathon/backend/internal/infrastructure/http/handler/report"
	visionhttp "cursor-hackathon/backend/internal/infrastructure/http/handler/vision"
	"cursor-hackathon/backend/internal/infrastructure/huggingface"
	"cursor-hackathon/backend/internal/infrastructure/openrouter"
	"cursor-hackathon/backend/internal/infrastructure/store"
	"cursor-hackathon/backend/internal/shared/config"
)

// NewMux builds the fully wired ServeMux (vision API + health).
func NewMux() *http.ServeMux {
	rules := config.MustLoad()
	pipeline := appvision.NewPipeline(rules)

	precomputed := demo.NewPrecomputedAdapter()
	roadDamage := demo.NewRoadDamageAdapter()

	router := appvision.NewRouter(precomputed)
	router.Register(domain.ModelModePrecomputed, precomputed)
	router.Register(domain.ModelModeRoadDamage, roadDamage)

	models := []visionhttp.ModelDescriptor{
		{ModelID: "civiclens/precomputed-detr-rdd", Mode: domain.ModelModePrecomputed, Role: "Reliable offline demo path", Live: false},
		{ModelID: demo.RoadDamageModelID, Mode: domain.ModelModeRoadDamage, Role: "Road damage (RDD2022) precomputed hero", Live: false},
	}

	// Live HF DETR is opt-in via token (graceful absence; design doc 12).
	var detr *huggingface.DETRAdapter
	if token := os.Getenv("HF_API_TOKEN"); token != "" {
		client := huggingface.NewClient(token, huggingface.WithBaseURL(os.Getenv("HF_INFERENCE_BASE_URL")))
		detr = huggingface.NewDETRAdapter(client)
		router.Register(domain.ModelModeLiveHF, detr)
		models = append(models, visionhttp.ModelDescriptor{
			ModelID: huggingface.DETRModelID, Mode: domain.ModelModeLiveHF, Role: "Live COCO baseline", Live: true,
		})
		log.Println("app: live HF DETR adapter registered")
	} else {
		log.Println("app: HF_API_TOKEN absent — precomputed-first, live HF disabled")
	}

	st := store.NewInMemory(200)

	// Reasoner is optional prose-only. With an OpenRouter key we rewrite the
	// summary via LLM but keep all decisions deterministic; otherwise the local
	// template reasoner is used (graceful degradation; design doc 8).
	var reasoner appvision.ReasonerPort = openrouter.NewLocalReasoner()
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		reasoner = openrouter.NewReasoner(key, os.Getenv("OPENROUTER_MODEL"), openrouter.NewLocalReasoner())
		log.Println("app: OpenRouter reasoner enabled (prose only)")
	}

	uc := appvision.NewAnalyzeImageUseCase(router, pipeline, st)

	handler := visionhttp.NewHandler(visionhttp.Deps{
		Analyze:  uc,
		Store:    st,
		Reasoner: reasoner,
		Models:   models,
		Modes:    router.Modes(),
	})

	// Wave 2 intake: KVKK anonymizer (HF detector if a token is present, else
	// no-op for synthetic/demo) -> vision -> dedup/route -> report store.
	var anon *anonymizer.Anonymizer
	if detr != nil {
		anon = anonymizer.New(anonymizer.NewHFDetector(detr), domain.PIIStrategyBlurApplied)
	} else {
		anon = anonymizer.New(anonymizer.NewNoopDetector(), domain.PIIStrategyAvoidanceByDesign)
	}
	reportStore := store.NewReportInMemory()
	intakeUC := intake.NewCreateReportUseCase(anonymizerAdapter{anon}, uc, reportStore, rules)
	reportHandler := reporthttp.NewHandler(intakeUC, reportStore)

	mux := http.NewServeMux()
	handler.Register(mux)
	reportHandler.Register(mux)
	registerHealth(mux)
	return mux
}

// anonymizerAdapter adapts the infrastructure anonymizer to the intake port.
type anonymizerAdapter struct{ inner *anonymizer.Anonymizer }

func (a anonymizerAdapter) Anonymize(ctx context.Context, img []byte) (intake.AnonymizationResult, error) {
	r, err := a.inner.Anonymize(ctx, img)
	if err != nil {
		return intake.AnonymizationResult{}, err
	}
	return intake.AnonymizationResult{
		Image:          r.Image,
		Width:          r.Width,
		Height:         r.Height,
		RegionsBlurred: r.RegionsBlurred,
		Strategy:       r.Strategy,
		Anonymized:     r.Anonymized,
	}, nil
}

func registerHealth(mux *http.ServeMux) {
	ok := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
	mux.HandleFunc("GET /health", ok)
	mux.HandleFunc("GET /health/live", ok)
	mux.HandleFunc("GET /health/ready", ok)
}
