package vision

import (
	"context"
	"time"

	domain "cursor-hackathon/backend/internal/domain/vision"
	"cursor-hackathon/backend/internal/shared/idgen"
)

// AnalyzeCommand is the input to the analyze use case. Image dimensions are
// resolved by the HTTP layer (it decodes the image) and passed in so the domain
// stays free of image-codec concerns.
type AnalyzeCommand struct {
	Image       []byte
	SourceType  string
	SourceRef   string
	ReportID    string // Wave 2: optional link to a Report (intake/completion)
	ModelMode   string
	ImageWidth  int
	ImageHeight int
	Location    *domain.Location
}

// AnalyzeImageUseCase orchestrates: resolve adapter -> infer -> run the
// deterministic pipeline -> build result -> persist -> publish event. It holds
// no decision logic itself; that lives in the pipeline.
type AnalyzeImageUseCase struct {
	router   ModelRouter
	pipeline *Pipeline
	store    AnalysisStorePort
	events   EventPublisherPort
	now      func() time.Time
	newID    func() string
}

// UseCaseOption configures the use case (DI for tests).
type UseCaseOption func(*AnalyzeImageUseCase)

// WithClock overrides the timestamp source.
func WithClock(now func() time.Time) UseCaseOption {
	return func(u *AnalyzeImageUseCase) { u.now = now }
}

// WithUseCaseIDFunc overrides the analysis ID generator.
func WithUseCaseIDFunc(fn func() string) UseCaseOption {
	return func(u *AnalyzeImageUseCase) { u.newID = fn }
}

// WithEventPublisher sets an optional async event sink.
func WithEventPublisher(ev EventPublisherPort) UseCaseOption {
	return func(u *AnalyzeImageUseCase) { u.events = ev }
}

// NewAnalyzeImageUseCase wires the use case. router, pipeline and store are
// required; events is optional.
func NewAnalyzeImageUseCase(router ModelRouter, pipeline *Pipeline, store AnalysisStorePort, opts ...UseCaseOption) *AnalyzeImageUseCase {
	u := &AnalyzeImageUseCase{
		router:   router,
		pipeline: pipeline,
		store:    store,
		now:      time.Now,
		newID:    idgen.NewUUID,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// Execute runs one analysis. The Kafka/event publish (if configured) happens
// after the result is built and persisted, and never fails the request.
func (u *AnalyzeImageUseCase) Execute(ctx context.Context, cmd AnalyzeCommand) (domain.AnalysisResult, error) {
	if len(cmd.Image) == 0 && cmd.SourceType != domain.SourceTypeSample {
		return domain.AnalysisResult{}, domain.ErrNoImage
	}

	mode := cmd.ModelMode
	if mode == "" {
		mode = domain.ModelModePrecomputed // precomputed-first default
	}

	adapter, err := u.router.Resolve(mode)
	if err != nil {
		return domain.AnalysisResult{}, err
	}

	raws, err := adapter.Detect(ctx, InferenceInput{
		Image:      cmd.Image,
		SourceType: cmd.SourceType,
		SourceRef:  cmd.SourceRef,
	})
	if err != nil {
		return domain.AnalysisResult{}, err
	}

	processed := u.pipeline.Process(raws)

	result := buildAnalysisResult(
		u.newID(),
		cmd,
		adapter.ModelID(),
		adapter.ModelMode(),
		processed,
		u.now().Format(time.RFC3339),
	)

	if err := u.store.Save(ctx, result); err != nil {
		return domain.AnalysisResult{}, err
	}

	if u.events != nil {
		u.events.PublishAnalysis(ctx, result)
	}

	return result, nil
}
