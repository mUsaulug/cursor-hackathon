package vision

import (
	"bytes"
	"image"
	// Register decoders so image.DecodeConfig can read dimensions.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"

	appvision "cursor-hackathon/backend/internal/application/vision"
	domain "cursor-hackathon/backend/internal/domain/vision"
)

const (
	maxUploadBytes  = 12 << 20 // 12 MiB cap on uploaded image
	defaultImageDim = 0
	sampleWidth     = 640 // fixtures are authored against a 640x480 canvas
	sampleHeight    = 480
)

// parseAnalyzeRequest builds an AnalyzeCommand from a multipart upload or a JSON
// sample request. The image is decoded only to read width/height (needed for
// bbox overlay); raw bytes are never persisted.
func parseAnalyzeRequest(r *http.Request) (appvision.AnalyzeCommand, error) {
	mode := r.URL.Query().Get("mode")

	contentType := r.Header.Get("Content-Type")
	if hasPrefix(contentType, "multipart/form-data") {
		return parseMultipart(r, mode)
	}
	return parseSample(r, mode)
}

func parseMultipart(r *http.Request, mode string) (appvision.AnalyzeCommand, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		return appvision.AnalyzeCommand{}, domain.ErrUnsupportedImage
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		return appvision.AnalyzeCommand{}, domain.ErrNoImage
	}
	defer file.Close()

	img, err := io.ReadAll(io.LimitReader(file, maxUploadBytes))
	if err != nil || len(img) == 0 {
		return appvision.AnalyzeCommand{}, domain.ErrNoImage
	}

	w, h := decodeDimensions(img)

	cmd := appvision.AnalyzeCommand{
		Image:       img,
		SourceType:  defaultStr(r.FormValue("source_type"), domain.SourceTypeUpload),
		SourceRef:   r.FormValue("source_ref"),
		ModelMode:   defaultStr(mode, r.FormValue("model_mode")),
		ImageWidth:  w,
		ImageHeight: h,
		Location:    parseLocation(r.FormValue("lat"), r.FormValue("lng")),
	}
	return cmd, nil
}

// parseSample handles a keyless demo request (no uploaded bytes): the caller
// picks a sample/precomputed scene via query params.
func parseSample(r *http.Request, mode string) (appvision.AnalyzeCommand, error) {
	q := r.URL.Query()
	ref := q.Get("source_ref")
	if ref == "" {
		ref = "sample_street_01"
	}
	return appvision.AnalyzeCommand{
		Image:       nil,
		SourceType:  domain.SourceTypeSample,
		SourceRef:   ref,
		ModelMode:   defaultStr(mode, domain.ModelModePrecomputed),
		ImageWidth:  sampleWidth,
		ImageHeight: sampleHeight,
		Location:    parseLocation(q.Get("lat"), q.Get("lng")),
	}, nil
}

func decodeDimensions(img []byte) (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(img))
	if err != nil {
		return defaultImageDim, defaultImageDim
	}
	return cfg.Width, cfg.Height
}

func parseLocation(latStr, lngStr string) *domain.Location {
	if latStr == "" || lngStr == "" {
		return nil
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		return nil
	}
	return &domain.Location{Lat: lat, Lng: lng}
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
