package middleware

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// RequestSizeLimit creates a middleware that limits the size of request bodies
// to prevent DoS attacks through oversized requests
func RequestSizeLimit(maxSize int64, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check Content-Length header first for efficiency
			if r.ContentLength > maxSize {
				logger.WithFields(logrus.Fields{
					"content_length": r.ContentLength,
					"max_size":       maxSize,
					"remote_addr":    r.RemoteAddr,
					"user_agent":     r.UserAgent(),
				}).Warn("Request rejected: Content-Length exceeds maximum size")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				w.Write([]byte(`{
					"error": "REQUEST_TOO_LARGE",
					"message": "Request body exceeds maximum allowed size",
					"max_size_bytes": ` + fmt.Sprintf("%d", maxSize) + `
				}`))
				return
			}

			// Wrap the request body with a size-limited reader
			// This protects against requests that don't set Content-Length
			// or set it incorrectly
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}
