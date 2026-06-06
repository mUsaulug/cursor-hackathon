package streetview

import (
	"context"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"testing"

	"cursor-hackathon/backend/internal/shared/imaging"
)

func sampleJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 4), uint8(y * 4), 128, 255})
		}
	}
	data, err := imaging.EncodeJPEG(img, 90)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return data
}

func TestFetchRequiresKey(t *testing.T) {
	s := NewSource("")
	if _, err := s.Fetch(context.Background(), 41.0, 28.9, ""); err == nil {
		t.Fatal("expected error without API key")
	}
}

func TestFetchAndAnonymize(t *testing.T) {
	jpg := sampleJPEG(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("location") == "" || r.URL.Query().Get("key") == "" {
			t.Errorf("missing query params: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(jpg)
	}))
	defer srv.Close()

	s := NewSource("test-key", WithBaseURL(srv.URL))
	raw, err := s.Fetch(context.Background(), 41.0082, 28.9784, "320x320")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("expected frame bytes")
	}

	anon, err := AnonymizeFrame(raw, []image.Rectangle{image.Rect(0, 0, 32, 32)}, 16)
	if err != nil {
		t.Fatalf("AnonymizeFrame: %v", err)
	}
	if _, _, err := imaging.Decode(anon); err != nil {
		t.Fatalf("anonymized frame not decodable: %v", err)
	}
}
