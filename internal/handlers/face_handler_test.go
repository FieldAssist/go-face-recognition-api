package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func newTestHandler() *FaceHandler {
	logger := logrus.New()
	return NewFaceHandler(nil, nil, nil, logger)
}

func TestDetectHandler_InvalidJSON(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/detect", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	h.DetectHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDetectHandler_InvalidURL(t *testing.T) {
	h := newTestHandler()
	body := []byte(`{"image_url":"ht!tp://bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/detect", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.DetectHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestValidateHandler_InvalidURL(t *testing.T) {
	h := newTestHandler()
	body := []byte(`{"image_url":"ht!tp://bad","min_faces":1,"max_faces":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/validate", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.ValidateHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDetectVisualHandler_InvalidURL(t *testing.T) {
	h := newTestHandler()
	body := []byte(`{"image_url":"ht!tp://bad","circle_color":"red","line_width":3}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/detect-visual", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.DetectVisualHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
