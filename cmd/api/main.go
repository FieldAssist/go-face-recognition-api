package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"face-recognition-api/internal/config"
	"face-recognition-api/internal/handlers"
	"face-recognition-api/internal/middleware"
	"face-recognition-api/internal/services"
	"face-recognition-api/internal/validation"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Initialize validation
	validation.Init()

	// Load configuration
	cfg := config.Load()

	logger.WithFields(logrus.Fields{
		"port":           cfg.Server.Port,
		"read_timeout":   cfg.Server.ReadTimeout,
		"write_timeout":  cfg.Server.WriteTimeout,
		"max_image_size": cfg.Limits.MaxImageSize,
	}).Info("Starting Face Recognition API")

	// Initialize services
	faceDetector, err := services.NewFaceDetector(cfg.Pigo, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize face detector")
	}

	optimizer := services.NewImageOptimizer(cfg.Optimizer, logger)
	imageDownloader := services.NewImageDownloader(cfg.Limits, optimizer, logger)
	imageProcessor := services.NewImageProcessor(logger)

	// Initialize handlers
	faceHandler := handlers.NewFaceHandler(faceDetector, imageDownloader, imageProcessor, logger)
	healthHandler := handlers.NewHealthHandler(logger)

	// Setup router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	// Security Fix: Add request size limit to prevent DoS attacks
	// Use MaxImageSize from config as the request size limit (typically 15MB)
	router.Use(middleware.RequestSizeLimit(int64(cfg.Limits.MaxImageSize), logger))

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Face detection endpoints
	api.HandleFunc("/detect", faceHandler.DetectHandler).Methods("POST")
	api.HandleFunc("/validate", faceHandler.ValidateHandler).Methods("POST")
	api.HandleFunc("/detect-visual", faceHandler.DetectVisualHandler).Methods("POST")

	// Health check endpoints
	api.HandleFunc("/health", healthHandler.HealthHandler).Methods("GET")
	api.HandleFunc("/ready", healthHandler.ReadinessHandler).Methods("GET")
	api.HandleFunc("/live", healthHandler.LivenessHandler).Methods("GET")

	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"face-recognition-api","version":"1.0.0","status":"running"}`))
	}).Methods("GET")

	// Configure HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown complete")
	}
}
