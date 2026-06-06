// Command anonymize irreversibly pixelates rectangular regions of a JPEG/PNG
// image (KVKK blur utility, design doc 5.3). It is a thin CLI over
// internal/shared/imaging so the blur can be demonstrated and verified on real
// images. WebP is not supported (stdlib limitation); convert to PNG/JPEG first.
//
// Usage:
//
//	go run ./cmd/anonymize -in raw.png -out blurred.png \
//	  -regions "x,y,w,h;x,y,w,h" -block 20
package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"

	"cursor-hackathon/backend/internal/shared/imaging"
)

func main() {
	in := flag.String("in", "", "input image path (jpeg/png)")
	out := flag.String("out", "", "output image path (.png or .jpg)")
	regionsArg := flag.String("regions", "", "regions to blur: x,y,w,h;x,y,w,h")
	block := flag.Int("block", 20, "mosaic block size in pixels")
	flag.Parse()

	if *in == "" || *out == "" || *regionsArg == "" {
		log.Fatal("anonymize: -in, -out and -regions are required")
	}

	data, err := os.ReadFile(*in)
	if err != nil {
		log.Fatalf("anonymize: read input: %v", err)
	}
	img, _, err := imaging.Decode(data)
	if err != nil {
		log.Fatalf("anonymize: %v (note: WebP is unsupported, convert to PNG/JPEG)", err)
	}

	regions, err := parseRegions(*regionsArg)
	if err != nil {
		log.Fatalf("anonymize: %v", err)
	}

	anon := imaging.PixelateRegions(img, regions, *block)

	var encoded []byte
	if strings.HasSuffix(strings.ToLower(*out), ".jpg") || strings.HasSuffix(strings.ToLower(*out), ".jpeg") {
		encoded, err = imaging.EncodeJPEG(anon, 88)
	} else {
		encoded, err = imaging.EncodePNG(anon)
	}
	if err != nil {
		log.Fatalf("anonymize: encode: %v", err)
	}
	if err := os.WriteFile(*out, encoded, 0o644); err != nil {
		log.Fatalf("anonymize: write output: %v", err)
	}
	fmt.Printf("anonymize: blurred %d region(s) -> %s\n", len(regions), *out)
}

func parseRegions(s string) ([]image.Rectangle, error) {
	var out []image.Rectangle
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		nums := strings.Split(part, ",")
		if len(nums) != 4 {
			return nil, fmt.Errorf("region %q must be x,y,w,h", part)
		}
		v := make([]int, 4)
		for i, n := range nums {
			x, err := strconv.Atoi(strings.TrimSpace(n))
			if err != nil {
				return nil, fmt.Errorf("region %q: %w", part, err)
			}
			v[i] = x
		}
		out = append(out, image.Rect(v[0], v[1], v[0]+v[2], v[1]+v[3]))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no regions parsed")
	}
	return out, nil
}
