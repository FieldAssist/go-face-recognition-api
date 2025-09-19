package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	// Import image decoders
	_ "image/jpeg"
	_ "image/png"

	"github.com/sirupsen/logrus"

	"face-recognition-api/internal/config"
	"face-recognition-api/internal/models"
)

// ImageDownloader handles downloading and validating images from URLs
type ImageDownloader struct {
	client    *http.Client
	config    config.LimitsConfig
	optimizer *ImageOptimizer
	logger    *logrus.Logger
}

// NewImageDownloader creates a new image downloader instance
func NewImageDownloader(cfg config.LimitsConfig, optimizer *ImageOptimizer, logger *logrus.Logger) *ImageDownloader {
	return &ImageDownloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// Security Fix: Enable proper TLS certificate validation
					InsecureSkipVerify: false,
					MinVersion:         tls.VersionTLS12, // Enforce minimum TLS 1.2
				},
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
			},
		},
		config:    cfg,
		optimizer: optimizer,
		logger:    logger,
	}
}

// DownloadImage downloads an image from the given URL and returns the decoded image
func (id *ImageDownloader) DownloadImage(ctx context.Context, imageURL string) (image.Image, models.ImageMetadata, error) {
	// Validate URL format
	if err := id.validateURL(imageURL); err != nil {
		return nil, models.ImageMetadata{}, fmt.Errorf("invalid URL: %w", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, models.ImageMetadata{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "Face-Recognition-API/1.0")
	req.Header.Set("Accept", "image/jpeg,image/png,image/*")

	// Execute request
	resp, err := id.client.Do(req)
	if err != nil {
		return nil, models.ImageMetadata{}, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, models.ImageMetadata{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !id.isValidImageType(contentType) {
		return nil, models.ImageMetadata{}, fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Check content length
	if resp.ContentLength > id.config.MaxImageSize {
		return nil, models.ImageMetadata{}, fmt.Errorf("image too large: %d bytes (max: %d)", resp.ContentLength, id.config.MaxImageSize)
	}

	// Decode image
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return nil, models.ImageMetadata{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Original metrics
	origBounds := img.Bounds()
	origW, origH := origBounds.Dx(), origBounds.Dy()
	finalSize := resp.ContentLength
	resized := false
	compressed := false
	jpegQ := 0

	// Optimize image (resize + compression) before detection
	if id.optimizer != nil {
		optImg, didResize, didCompress, usedQ, encSize, optErr := id.optimizer.Optimize(img)
		if optErr == nil {
			img = optImg
			resized = didResize
			compressed = didCompress
			jpegQ = usedQ
			if encSize > 0 {
				finalSize = encSize
			}
		} else {
			id.logger.WithError(optErr).Warn("image optimization failed; proceeding with original image")
		}
	}

	// Final dimensions
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Create metadata
	metadata := models.ImageMetadata{
		Width:             width,
		Height:            height,
		Format:            strings.ToUpper(format),
		SizeBytes:         finalSize,
		URL:               imageURL,
		Optimized:         resized || compressed,
		Resized:           resized,
		OriginalWidth:     origW,
		OriginalHeight:    origH,
		OriginalSizeBytes: resp.ContentLength,
		JpegQualityUsed:   jpegQ,
	}

	id.logger.WithFields(logrus.Fields{
		"url":          imageURL,
		"width":        width,
		"height":       height,
		"format":       format,
		"size_bytes":   finalSize,
		"resized":      resized,
		"compressed":   compressed,
		"jpeg_quality": jpegQ,
	}).Info("Image downloaded and optimized successfully")

	return img, metadata, nil
}

// validateURL validates the image URL format and security
func (id *ImageDownloader) validateURL(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Check host
	if parsedURL.Host == "" {
		return fmt.Errorf("missing host in URL")
	}

	// Security Fix: Enhanced SSRF protection - block private IP ranges
	isPrivate, err := id.isPrivateIP(parsedURL.Host)
	if err != nil {
		return fmt.Errorf("failed to validate host: %w", err)
	}
	if isPrivate {
		return fmt.Errorf("access to private IP ranges is not allowed")
	}

	return nil
}

// isValidImageType checks if the content type is a supported image format
func (id *ImageDownloader) isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	contentType = strings.ToLower(strings.Split(contentType, ";")[0])
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}

	return false
}

// isPrivateIP performs comprehensive check for private IP ranges using Go's net package
// Security Fix: Proper SSRF protection with IPv4/IPv6 support and DNS resolution
func (id *ImageDownloader) isPrivateIP(host string) (bool, error) {
	// First, try to parse as IP address directly
	if addr, err := netip.ParseAddr(host); err == nil {
		// Successfully parsed as IP address
		return id.isPrivateAddr(addr), nil
	}

	// If not a direct IP, try to resolve hostname to IP addresses
	ips, err := net.LookupIP(host)
	if err != nil {
		return false, fmt.Errorf("failed to resolve host %s: %w", host, err)
	}

	if len(ips) == 0 {
		return false, fmt.Errorf("no IP addresses found for host: %s", host)
	}

	// Check all resolved IP addresses - if any are private, block the request
	for _, ip := range ips {
		if addr, ok := netip.AddrFromSlice(ip); ok {
			if id.isPrivateAddr(addr) {
				id.logger.WithFields(logrus.Fields{
					"host":        host,
					"resolved_ip": addr.String(),
				}).Warn("Blocked request to private IP address")
				return true, nil
			}
		}
	}

	return false, nil
}

// isPrivateAddr checks if an IP address is in private ranges
func (id *ImageDownloader) isPrivateAddr(addr netip.Addr) bool {
	return addr.IsPrivate() ||
		addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified()
}
