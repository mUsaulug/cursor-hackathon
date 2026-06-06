// Package imaging provides irreversible, in-memory region anonymization for
// KVKK compliance (design doc 5.3; docs/PRIVACY.md). It uses only the standard
// library (image, image/draw, image/jpeg, image/png). NOTE: the stdlib cannot
// decode WebP; Street View and uploads deliver JPEG/PNG, which is sufficient.
//
// Anonymization here is pixelation: each PII region is replaced by averaged
// blocks. This is destructive and cannot be reversed, which is exactly what
// KVKK requires — we blur to anonymize, never to read or identify.
package imaging

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"

	domain "cursor-hackathon/backend/internal/domain/vision"
)

// Decode reads a JPEG or PNG image. WebP is intentionally unsupported (stdlib).
func Decode(data []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("imaging: decode: %w", err)
	}
	return img, format, nil
}

// EncodeJPEG encodes an image to JPEG bytes.
func EncodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("imaging: encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

// EncodePNG encodes an image to PNG bytes.
func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("imaging: encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// BBoxToRect converts a domain bounding box to an integer image rectangle.
func BBoxToRect(b domain.BoundingBox) image.Rectangle {
	return image.Rect(int(b.XMin), int(b.YMin), int(b.XMax), int(b.YMax))
}

// PixelateRegions returns a copy of src with each region irreversibly pixelated.
// block is the mosaic block size in pixels (larger => coarser/more anonymized);
// values < 1 are clamped to a safe default.
func PixelateRegions(src image.Image, regions []image.Rectangle, block int) *image.RGBA {
	if block < 1 {
		block = 16
	}
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)

	for _, region := range regions {
		r := region.Intersect(b)
		if r.Empty() {
			continue
		}
		pixelateBlock(dst, r, block)
	}
	return dst
}

// pixelateBlock replaces each block within r by its average color.
func pixelateBlock(dst *image.RGBA, r image.Rectangle, block int) {
	for by := r.Min.Y; by < r.Max.Y; by += block {
		for bx := r.Min.X; bx < r.Max.X; bx += block {
			cell := image.Rect(bx, by, bx+block, by+block).Intersect(r)
			ar, ag, ab, aa := averageColor(dst, cell)
			fillRect(dst, cell, ar, ag, ab, aa)
		}
	}
}

func averageColor(img *image.RGBA, r image.Rectangle) (uint8, uint8, uint8, uint8) {
	var sr, sg, sb, sa, n uint64
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			i := img.PixOffset(x, y)
			sr += uint64(img.Pix[i])
			sg += uint64(img.Pix[i+1])
			sb += uint64(img.Pix[i+2])
			sa += uint64(img.Pix[i+3])
			n++
		}
	}
	if n == 0 {
		return 0, 0, 0, 255
	}
	return uint8(sr / n), uint8(sg / n), uint8(sb / n), uint8(sa / n)
}

func fillRect(img *image.RGBA, r image.Rectangle, cr, cg, cb, ca uint8) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = cr
			img.Pix[i+1] = cg
			img.Pix[i+2] = cb
			img.Pix[i+3] = ca
		}
	}
}
