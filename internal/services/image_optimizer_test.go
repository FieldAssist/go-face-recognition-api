package services

import (
	"image"
	"image/color"
	"testing"

	"github.com/sirupsen/logrus"

	"face-recognition-api/internal/config"
)

func TestResizeToMaxSide(t *testing.T) {
	logger := logrus.New()
	opt := NewImageOptimizer(config.OptimizerConfig{MaxSide: 1024}, logger)

	// Create a large test image 3000x2000
	src := image.NewRGBA(image.Rect(0, 0, 3000, 2000))
	img, resized := opt.ResizeToMaxSide(src, 1024)
	if !resized {
		t.Fatalf("expected resized to be true")
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > 1024 || h > 1024 {
		t.Fatalf("expected both sides <= 1024, got %dx%d", w, h)
	}
}

func TestOptimize_CompressTarget(t *testing.T) {
	logger := logrus.New()
	opt := NewImageOptimizer(config.OptimizerConfig{
		MaxSide:        0, // no resize
		TargetMaxBytes: 50 * 1024,
		JPEGMinQuality: 50,
		JPEGMaxQuality: 90,
	}, logger)

	// Create a 512x512 color test image
	src := image.NewRGBA(image.Rect(0, 0, 512, 512))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			src.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 128, A: 255})
		}
	}

	img, resized, compressed, q, size, err := opt.Optimize(src)
	if err != nil {
		t.Fatalf("optimize error: %v", err)
	}
	_ = img
	if resized {
		t.Fatalf("did not expect resize")
	}
	if !compressed {
		t.Fatalf("expected compression to occur")
	}
	if size <= 0 || size > int64(50*1024) {
		t.Fatalf("expected size <= 50KB, got %d", size)
	}
	if q < 50 || q > 90 {
		t.Fatalf("unexpected jpeg quality %d", q)
	}
}

