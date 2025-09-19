package models

// FaceDetectionRequest represents the request for face detection endpoint
type FaceDetectionRequest struct {
	ImageURL string `json:"image_url" binding:"required,url" validate:"required,url"`
}

// SelfieValidationRequest represents the request for selfie validation endpoint
type SelfieValidationRequest struct {
	ImageURL string `json:"image_url" binding:"required,url" validate:"required,url"`
	MinFaces int    `json:"min_faces" default:"1" validate:"omitempty,min=0"`
	MaxFaces int    `json:"max_faces" default:"1" validate:"omitempty,min=0"`
}

// VisualDetectionRequest represents the request for visual detection endpoint
type VisualDetectionRequest struct {
	ImageURL    string `json:"image_url" binding:"required,url" validate:"required,url"`
	CircleColor string `json:"circle_color" default:"red" validate:"omitempty,alpha"`
	LineWidth   int    `json:"line_width" default:"3" validate:"omitempty,min=1,max=20"`
}
