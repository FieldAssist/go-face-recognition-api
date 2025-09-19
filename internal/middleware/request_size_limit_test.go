package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestRequestSizeLimit_ContentLengthTooLarge(t *testing.T) {
	logger := logrus.New()
	mw := RequestSizeLimit(10, logger) // 10 bytes max

	// Handler should not be called when Content-Length exceeds
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	h := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789012345")) // 16 bytes
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if called {
		t.Fatalf("next handler should not be called")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestRequestSizeLimit_AllowsWithinLimit(t *testing.T) {
	logger := logrus.New()
	mw := RequestSizeLimit(10, logger) // 10 bytes max

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body to ensure MaxBytesReader allows within limit
		_, _ = io.ReadAll(r.Body)
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789")) // 10 bytes
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

