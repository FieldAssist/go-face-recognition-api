package models

// Face represents a detected face with coordinates and confidence
type Face struct {
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Confidence float32 `json:"confidence"`
}

// ImageMetadata contains metadata about the processed image
type ImageMetadata struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	SizeBytes int64  `json:"size_bytes"`
	URL       string `json:"url"`
	// Optimization details (optional)
	Optimized         bool  `json:"optimized,omitempty"`
	Resized           bool  `json:"resized,omitempty"`
	OriginalWidth     int   `json:"original_width,omitempty"`
	OriginalHeight    int   `json:"original_height,omitempty"`
	OriginalSizeBytes int64 `json:"original_size_bytes,omitempty"`
	JpegQualityUsed   int   `json:"jpeg_quality_used,omitempty"`
}

// FaceDetectionResponse represents the response for face detection endpoint
type FaceDetectionResponse struct {
	Faces            []Face        `json:"faces"`
	Count            int           `json:"count"`
	ImageMetadata    ImageMetadata `json:"image_metadata"`
	ProcessingTimeMs float64       `json:"processing_time_ms"`
}

// SelfieValidationResponse represents the response for selfie validation endpoint
type SelfieValidationResponse struct {
	IsValid    bool     `json:"is_valid"`
	Issues     []string `json:"issues,omitempty"`
	Confidence float32  `json:"confidence"`
	FaceCount  int      `json:"face_count"`
}

// VisualDetectionResponse represents the response for visual detection endpoint
type VisualDetectionResponse struct {
	ImageBase64      string        `json:"image_base64"`
	Faces            []Face        `json:"faces"`
	Count            int           `json:"count"`
	ImageMetadata    ImageMetadata `json:"image_metadata"`
	ProcessingTimeMs float64       `json:"processing_time_ms"`
}

// APIError represents a structured API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// Predefined API errors
var (
	ErrInvalidURL     = &APIError{Code: "INVALID_URL", Message: "Invalid image URL", Status: 400}
	ErrImageDownload  = &APIError{Code: "IMAGE_DOWNLOAD_ERROR", Message: "Failed to download image", Status: 400}
	ErrImageFormat    = &APIError{Code: "INVALID_IMAGE_FORMAT", Message: "Unsupported image format", Status: 400}
	ErrFaceDetection  = &APIError{Code: "FACE_DETECTION_ERROR", Message: "Face detection failed", Status: 500}
	ErrImageTooLarge  = &APIError{Code: "IMAGE_TOO_LARGE", Message: "Image size exceeds maximum limit", Status: 400}
	ErrInvalidRequest = &APIError{Code: "INVALID_REQUEST", Message: "Invalid JSON request", Status: 400}
)
