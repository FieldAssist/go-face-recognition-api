package services

import (
	_ "embed"
	"fmt"
	"image"

	pigo "github.com/esimov/pigo/core"
	"github.com/sirupsen/logrus"

	"face-recognition-api/internal/config"
	"face-recognition-api/internal/models"
)

// Embed the cascade file - you'll need to download this from pigo repository
//
//go:embed facefinder
var cascadeFile []byte

// FaceDetector wraps the pigo face detection library
type FaceDetector struct {
	classifier *pigo.Pigo
	config     config.PigoConfig
	logger     *logrus.Logger
}

// NewFaceDetector creates a new face detector instance
func NewFaceDetector(cfg config.PigoConfig, logger *logrus.Logger) (*FaceDetector, error) {
	// Initialize pigo classifier
	p := pigo.NewPigo()

	// Parse the cascade file
	classifier, err := p.Unpack(cascadeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cascade file: %w", err)
	}

	return &FaceDetector{
		classifier: classifier,
		config:     cfg,
		logger:     logger,
	}, nil
}

// DetectFaces detects faces in the given image and returns face coordinates
func (fd *FaceDetector) DetectFaces(img image.Image) ([]models.Face, error) {
	// Security Fix: Add nil pointer validation to prevent panics
	if img == nil {
		return nil, fmt.Errorf("image cannot be nil")
	}

	// Get image dimensions with proper bounds calculation
	bounds := img.Bounds()
	width := bounds.Dx()  // Use Dx() instead of Max.X for proper width calculation
	height := bounds.Dy() // Use Dy() instead of Max.Y for proper height calculation

	// Validate image dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	// Convert image to grayscale using pigo's utility
	pixels := pigo.RgbToGrayscale(img)

	// Set up cascade parameters
	cParams := pigo.CascadeParams{
		MinSize:     fd.config.MinSize,
		MaxSize:     fd.config.MaxSize,
		ShiftFactor: float64(fd.config.ShiftFactor),
		ScaleFactor: float64(fd.config.ScaleFactor),
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   height,
			Cols:   width,
			Dim:    width,
		},
	}

	// Run face detection
	angle := 0.0 // No rotation
	detections := fd.classifier.RunCascade(cParams, angle)

	// Cluster detections to remove duplicates
	detections = fd.classifier.ClusterDetections(detections, float64(fd.config.IoUThreshold))

	// Filter detections by confidence threshold
	var filteredDetections []pigo.Detection
	for _, det := range detections {
		if float32(det.Q) >= fd.config.MinConfidence {
			filteredDetections = append(filteredDetections, det)
		}
	}

	// Convert to our Face model
	faces := make([]models.Face, len(filteredDetections))
	for i, det := range filteredDetections {
		faces[i] = models.Face{
			X:          det.Col - det.Scale/2,
			Y:          det.Row - det.Scale/2,
			Width:      det.Scale,
			Height:     det.Scale,
			Confidence: det.Q,
		}
	}

	return faces, nil
}

// ValidateSelfie validates if the image is a good selfie based on face count and quality
func (fd *FaceDetector) ValidateSelfie(faces []models.Face, minFaces, maxFaces int) models.SelfieValidationResponse {
	faceCount := len(faces)
	issues := make([]string, 0)
	isValid := true
	var confidence float32

	// Check face count
	if faceCount < minFaces {
		isValid = false
		if faceCount == 0 {
			issues = append(issues, "No faces detected in image")
			issues = append(issues, "Image may be too dark or blurry")
		} else {
			issues = append(issues, fmt.Sprintf("Too few faces detected (%d found, expected at least %d)", faceCount, minFaces))
		}
	} else if faceCount > maxFaces {
		isValid = false
		issues = append(issues, fmt.Sprintf("Multiple faces detected (%d found, expected %d)", faceCount, maxFaces))
	}

	// Calculate confidence (average of all face confidences)
	if faceCount > 0 {
		var totalConfidence float32
		for _, face := range faces {
			totalConfidence += face.Confidence
		}
		confidence = totalConfidence / float32(faceCount)

		// Check confidence threshold
		if confidence < 10.0 {
			isValid = false
			issues = append(issues, "Low confidence score for detected face(s)")
		}
	}

	return models.SelfieValidationResponse{
		IsValid:    isValid,
		Issues:     issues,
		Confidence: confidence,
		FaceCount:  faceCount,
	}
}
