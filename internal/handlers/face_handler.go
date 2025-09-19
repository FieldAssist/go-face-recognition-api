package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"face-recognition-api/internal/models"
	"face-recognition-api/internal/services"
	"face-recognition-api/internal/validation"
)

// FaceHandler handles face detection related endpoints
type FaceHandler struct {
	faceDetector    *services.FaceDetector
	imageDownloader *services.ImageDownloader
	imageProcessor  *services.ImageProcessor
	logger          *logrus.Logger
}

// NewFaceHandler creates a new face handler instance
func NewFaceHandler(
	fd *services.FaceDetector,
	id *services.ImageDownloader,
	ip *services.ImageProcessor,
	logger *logrus.Logger,
) *FaceHandler {
	return &FaceHandler{
		faceDetector:    fd,
		imageDownloader: id,
		imageProcessor:  ip,
		logger:          logger,
	}
}

// DetectHandler handles POST /api/v1/detect endpoint
func (h *FaceHandler) DetectHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req models.FaceDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request", err)
		return
	}

	// Validate request using validator
	if err := validation.ValidateStruct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation error: "+err.Error(), err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Download image
	img, metadata, err := h.imageDownloader.DownloadImage(ctx, req.ImageURL)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "IMAGE_DOWNLOAD_FAILED", "Failed to download image", err)
		return
	}

	// Detect faces
	faces, err := h.faceDetector.DetectFaces(img)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "FACE_DETECTION_FAILED", "Face detection failed", err)
		return
	}

	processingTime := time.Since(start).Seconds() * 1000

	response := models.FaceDetectionResponse{
		Faces:            faces,
		Count:            len(faces),
		ImageMetadata:    metadata,
		ProcessingTimeMs: processingTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateHandler handles POST /api/v1/validate endpoint
func (h *FaceHandler) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req models.SelfieValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request", err)
		return
	}

	// Validate request using validator
	if err := validation.ValidateStruct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation error: "+err.Error(), err)
		return
	}

	// Set defaults
	if req.MinFaces == 0 {
		req.MinFaces = 1
	}
	if req.MaxFaces == 0 {
		req.MaxFaces = 1
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Download image
	img, _, err := h.imageDownloader.DownloadImage(ctx, req.ImageURL)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "IMAGE_DOWNLOAD_FAILED", "Failed to download image", err)
		return
	}

	// Detect faces
	faces, err := h.faceDetector.DetectFaces(img)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "FACE_DETECTION_FAILED", "Face detection failed", err)
		return
	}

	// Validate selfie
	response := h.faceDetector.ValidateSelfie(faces, req.MinFaces, req.MaxFaces)

	h.logger.WithFields(logrus.Fields{
		"url":             req.ImageURL,
		"faces_detected":  len(faces),
		"is_valid":        response.IsValid,
		"processing_time": time.Since(start).Milliseconds(),
	}).Info("Selfie validation completed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DetectVisualHandler handles POST /api/v1/detect-visual endpoint
func (h *FaceHandler) DetectVisualHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req models.VisualDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request", err)
		return
	}

	// Validate request using validator
	if err := validation.ValidateStruct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation error: "+err.Error(), err)
		return
	}

	// Set defaults
	if req.CircleColor == "" {
		req.CircleColor = "red"
	}
	if req.LineWidth == 0 {
		req.LineWidth = 3
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Download image
	img, metadata, err := h.imageDownloader.DownloadImage(ctx, req.ImageURL)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "IMAGE_DOWNLOAD_FAILED", "Failed to download image", err)
		return
	}

	// Detect faces
	faces, err := h.faceDetector.DetectFaces(img)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "FACE_DETECTION_FAILED", "Face detection failed", err)
		return
	}

	// Parse color and create circle options
	circleColor := h.imageProcessor.ParseColor(req.CircleColor)
	circleOpts := services.CircleOptions{
		Color:     circleColor,
		LineWidth: req.LineWidth,
	}

	// Draw circles on image
	imageBase64, err := h.imageProcessor.DrawFaceCircles(img, faces, circleOpts)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "IMAGE_PROCESSING_FAILED", "Failed to process image", err)
		return
	}

	processingTime := time.Since(start).Seconds() * 1000

	response := models.VisualDetectionResponse{
		ImageBase64:      imageBase64,
		Faces:            faces,
		Count:            len(faces),
		ImageMetadata:    metadata,
		ProcessingTimeMs: processingTime,
	}

	h.logger.WithFields(logrus.Fields{
		"url":             req.ImageURL,
		"faces_detected":  len(faces),
		"circle_color":    req.CircleColor,
		"line_width":      req.LineWidth,
		"processing_time": processingTime,
	}).Info("Visual detection completed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes a structured error response
func (h *FaceHandler) writeErrorResponse(w http.ResponseWriter, status int, code, message string, err error) {
	if err != nil {
		h.logger.WithError(err).Error(message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResp := map[string]interface{}{
		"error": message,
		"code":  code,
	}

	json.NewEncoder(w).Encode(errorResp)
}
