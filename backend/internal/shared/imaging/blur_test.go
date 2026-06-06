package imaging

import (
	"image"
	"image/color"
	"testing"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// makeImage builds an RGBA image filled with a checkerboard so a pixelated
// region becomes detectably uniform.
func makeImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			} else {
				img.Set(x, y, color.RGBA{0, 0, 255, 255})
			}
		}
	}
	return img
}

func TestPixelateRegionsChangesOnlyRegion(t *testing.T) {
	src := makeImage(64, 64)
	region := image.Rect(10, 10, 40, 40)
	out := PixelateRegions(src, []image.Rectangle{region}, 8)

	// Inside the region, a full block must be uniform (pixelated).
	r0, g0, b0, _ := out.At(10, 10).RGBA()
	uniform := true
	for y := 10; y < 18; y++ {
		for x := 10; x < 18; x++ {
			r, g, b, _ := out.At(x, y).RGBA()
			if r != r0 || g != g0 || b != b0 {
				uniform = false
			}
		}
	}
	if !uniform {
		t.Error("expected the first block inside the region to be uniform after pixelation")
	}

	// Outside the region, pixels must be untouched (still checkerboard).
	if out.At(0, 0) != src.At(0, 0) {
		t.Error("pixels outside the region must be unchanged")
	}
	if out.At(50, 50) != src.At(50, 50) {
		t.Error("pixels outside the region must be unchanged")
	}
}

func TestPixelateIsIrreversibleAveraging(t *testing.T) {
	// A region over a 50/50 red/blue checkerboard should average to purple-ish,
	// losing the original alternating values (irreversible).
	src := makeImage(16, 16)
	out := PixelateRegions(src, []image.Rectangle{image.Rect(0, 0, 16, 16)}, 16)
	r, g, b, _ := out.At(8, 8).RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	if r8 == 255 && b8 == 0 || r8 == 0 && b8 == 255 {
		t.Errorf("expected averaged color, got pure (%d,%d,%d)", r8, g8, b8)
	}
	_ = g8
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	src := makeImage(32, 32)
	jpg, err := EncodeJPEG(src, 90)
	if err != nil {
		t.Fatalf("EncodeJPEG: %v", err)
	}
	if _, format, err := Decode(jpg); err != nil || format != "jpeg" {
		t.Errorf("decode jpeg: format=%q err=%v", format, err)
	}

	pngBytes, err := EncodePNG(src)
	if err != nil {
		t.Fatalf("EncodePNG: %v", err)
	}
	if _, format, err := Decode(pngBytes); err != nil || format != "png" {
		t.Errorf("decode png: format=%q err=%v", format, err)
	}
}

func TestBBoxToRect(t *testing.T) {
	r := BBoxToRect(domain.BoundingBox{XMin: 5, YMin: 6, XMax: 20, YMax: 30})
	if r != image.Rect(5, 6, 20, 30) {
		t.Errorf("BBoxToRect = %v", r)
	}
}
