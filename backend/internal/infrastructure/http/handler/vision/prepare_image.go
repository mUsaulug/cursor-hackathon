package vision

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

var errStreetViewUnavailable = errors.New("streetview: API key not configured")

func (h *Handler) prepareImage(ctx context.Context, r *http.Request, cmd *appvision.AnalyzeCommand) error {
	if cmd.SourceRef == "streetview" {
		return h.prepareStreetView(ctx, r, cmd)
	}
	if len(cmd.Image) > 0 && cmd.SourceType != domain.SourceTypeSample {
		return h.anonymizeUpload(ctx, cmd)
	}
	return nil
}

func (h *Handler) prepareStreetView(ctx context.Context, r *http.Request, cmd *appvision.AnalyzeCommand) error {
	if h.streetView == nil {
		return errStreetViewUnavailable
	}
	q := r.URL.Query()
	lat, err1 := strconv.ParseFloat(q.Get("lat"), 64)
	lng, err2 := strconv.ParseFloat(q.Get("lng"), 64)
	if err1 != nil || err2 != nil {
		return domain.ErrNoImage
	}
	raw, err := h.streetView.Fetch(ctx, lat, lng)
	if err != nil {
		return fmt.Errorf("streetview fetch: %w", err)
	}
	if h.anonymizer == nil {
		return errors.New("anonymizer not configured")
	}
	anon, err := h.anonymizer.Anonymize(ctx, raw)
	if err != nil {
		return err
	}
	cmd.Image = anon.Image
	cmd.ImageWidth = anon.Width
	cmd.ImageHeight = anon.Height
	cmd.SourceType = domain.SourceTypeStreetView
	cmd.AnonymizationMeta = &appvision.AnonymizationMeta{
		Strategy:       anon.Strategy,
		Anonymized:     anon.Anonymized,
		RegionsBlurred: anon.RegionsBlurred,
	}
	return nil
}

func (h *Handler) anonymizeUpload(ctx context.Context, cmd *appvision.AnalyzeCommand) error {
	if h.anonymizer == nil {
		return errors.New("anonymizer not configured")
	}
	anon, err := h.anonymizer.Anonymize(ctx, cmd.Image)
	if err != nil {
		return err
	}
	cmd.Image = anon.Image
	cmd.ImageWidth = anon.Width
	cmd.ImageHeight = anon.Height
	cmd.AnonymizationMeta = &appvision.AnonymizationMeta{
		Strategy:       anon.Strategy,
		Anonymized:     anon.Anonymized,
		RegionsBlurred: anon.RegionsBlurred,
	}
	return nil
}
