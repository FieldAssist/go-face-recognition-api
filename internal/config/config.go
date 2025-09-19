package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Pigo      PigoConfig
	Limits    LimitsConfig
	Optimizer OptimizerConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// PigoConfig holds pigo face detection configuration
type PigoConfig struct {
	MinSize       int
	MaxSize       int
	ShiftFactor   float32
	ScaleFactor   float32
	IoUThreshold  float32
	MinConfidence float32
}

// LimitsConfig holds various limits for the application
type LimitsConfig struct {
	MaxImageSize int64
	MaxWidth     int
	MaxHeight    int
}

// OptimizerConfig holds image optimization settings
type OptimizerConfig struct {
	MaxSide        int   // Longest image side in pixels before detection
	TargetMaxBytes int64 // Target maximum encoded size in bytes (approximate)
	JPEGMinQuality int   // Minimum JPEG quality used during adaptive compression
	JPEGMaxQuality int   // Maximum JPEG quality used during adaptive compression
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", ":8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 120*time.Second),
		},
		Pigo: PigoConfig{
			MinSize:       getIntEnv("PIGO_MIN_SIZE", 20),
			MaxSize:       getIntEnv("PIGO_MAX_SIZE", 1000),
			ShiftFactor:   getFloat32Env("PIGO_SHIFT_FACTOR", 0.1),
			ScaleFactor:   getFloat32Env("PIGO_SCALE_FACTOR", 1.1),
			IoUThreshold:  getFloat32Env("PIGO_IOU_THRESHOLD", 0.2),
			MinConfidence: getFloat32Env("PIGO_MIN_CONFIDENCE", 5.0),
		},
		Limits: LimitsConfig{
			MaxImageSize: getInt64Env("MAX_IMAGE_SIZE", 15728640), // 15MB
			MaxWidth:     getIntEnv("MAX_WIDTH", 5000),
			MaxHeight:    getIntEnv("MAX_HEIGHT", 5000),
		},
		Optimizer: OptimizerConfig{
			MaxSide:        getIntEnv("OPT_MAX_SIDE", 1024),
			TargetMaxBytes: getInt64Env("OPT_TARGET_MAX_BYTES", 1000000),
			JPEGMinQuality: getIntEnv("OPT_JPEG_MIN_QUALITY", 60),
			JPEGMaxQuality: getIntEnv("OPT_JPEG_MAX_QUALITY", 90),
		},
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFloat32Env(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
