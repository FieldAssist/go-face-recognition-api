package interfaces

import (
	"context"
	"image"

	"face-recognition-api/internal/models"
)

// FaceDetector defines face detection capabilities
type FaceDetector interface {
	DetectFaces(img image.Image) ([]models.Face, error)
	ValidateSelfie(faces []models.Face, minFaces, maxFaces int) models.SelfieValidationResponse
}

// ImageDownloader defines image download and validation behavior
type ImageDownloader interface {
	DownloadImage(ctx context.Context, imageURL string) (image.Image, models.ImageMetadata, error)
}
