package services

import (
	"bytes"
	"image"
	"image/color"
	imagedraw "image/draw"
	"image/jpeg"

	"github.com/sirupsen/logrus"
	xdraw "golang.org/x/image/draw"

	"face-recognition-api/internal/config"
)

// ImageOptimizer provides image resizing and adaptive compression.
// It is used to reduce image dimensions and size before face detection
// to improve performance while preserving detection quality.
type ImageOptimizer struct {
	cfg    config.OptimizerConfig
	logger *logrus.Logger
}

// NewImageOptimizer constructs a new optimizer with config and logger.
func NewImageOptimizer(cfg config.OptimizerConfig, logger *logrus.Logger) *ImageOptimizer {
	return &ImageOptimizer{cfg: cfg, logger: logger}
}

// ResizeToMaxSide resizes the image so that the longest side is <= maxSide, keeping aspect ratio.
// Returns the possibly resized image and a flag indicating if a resize was applied.
func (o *ImageOptimizer) ResizeToMaxSide(src image.Image, maxSide int) (image.Image, bool) {
	if src == nil {
		return nil, false
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if maxSide <= 0 {
		return src, false
	}
	long := w
	if h > long {
		long = h
	}
	if long <= maxSide {
		return src, false
	}

	scale := float64(maxSide) / float64(long)
	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	// Use high-quality resampler (CatmullRom)
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, b, imagedraw.Over, nil)
	return dst, true
}

// flattenToOpaque paints the source image onto a white background to remove alpha channel
// before JPEG encoding (which does not support alpha).
func flattenToOpaque(src image.Image) *image.RGBA {
	b := src.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	imagedraw.Draw(rgba, rgba.Bounds(), &image.Uniform{color.White}, image.Point{}, imagedraw.Src)
	imagedraw.Draw(rgba, rgba.Bounds(), src, b.Min, imagedraw.Over)
	return rgba
}

// compressToTargetJPEG adaptively encodes the image to JPEG within the target size, if possible.
// Returns: encoded image decoded back, used quality, final size, compressed flag.
func (o *ImageOptimizer) compressToTargetJPEG(src image.Image, target int64) (image.Image, int, int64, bool, error) {
	if target <= 0 {
		return src, 0, 0, false, nil
	}

	// Ensure opaque background for JPEG
	opaque := flattenToOpaque(src)

	minQ := o.cfg.JPEGMinQuality
	maxQ := o.cfg.JPEGMaxQuality
	if minQ <= 0 {
		minQ = 60
	}
	if maxQ <= 0 {
		maxQ = 90
	}
	if maxQ < minQ {
		maxQ = minQ
	}

	var bestImg image.Image
	var bestQ int
	var bestSize int64
	compressed := false

	// Try from high quality downwards to find the highest quality under target.
	for q := maxQ; q >= minQ; q -= 5 {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, opaque, &jpeg.Options{Quality: q}); err != nil {
			return src, 0, 0, false, err
		}
		sz := int64(buf.Len())
		if sz <= target {
			// Decode back to image.Image for downstream processing (ensures memory image used by detector)
			img, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
			if err != nil {
				return src, 0, 0, false, err
			}
			bestImg = img
			bestQ = q
			bestSize = sz
			compressed = true
			break
		}
	}

	if compressed {
		return bestImg, bestQ, bestSize, true, nil
	}
	return src, 0, 0, false, nil
}

// Optimize performs resize to MaxSide and adaptive JPEG compression to TargetMaxBytes.
// Returns the optimized image and flags/metrics.
func (o *ImageOptimizer) Optimize(src image.Image) (out image.Image, resized bool, compressed bool, usedQuality int, finalSize int64, err error) {
	if src == nil {
		return nil, false, false, 0, 0, nil
	}

	out = src
	if o.cfg.MaxSide > 0 {
		if r, ok := o.ResizeToMaxSide(out, o.cfg.MaxSide); ok {
			out = r
			resized = true
		}
	}

	if o.cfg.TargetMaxBytes > 0 {
		img2, q, size, ok, e := o.compressToTargetJPEG(out, o.cfg.TargetMaxBytes)
		if e != nil {
			return out, resized, false, 0, 0, e
		}
		if ok {
			out = img2
			compressed = true
			usedQuality = q
			finalSize = size
		}
	}
	return out, resized, compressed, usedQuality, finalSize, nil
}
