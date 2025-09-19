package services

import (
	"testing"
	"face-recognition-api/internal/config"
)

// TestDetectFaces_NilImage ensures nil images are rejected to prevent panics
func TestDetectFaces_NilImage(t *testing.T) {
	fd := &FaceDetector{config: config.PigoConfig{}}
	faces, err := fd.DetectFaces(nil)
	if err == nil {
		t.Fatalf("expected error for nil image, got nil")
	}
	if faces != nil {
		t.Fatalf("expected nil faces on error, got %#v", faces)
	}
}

