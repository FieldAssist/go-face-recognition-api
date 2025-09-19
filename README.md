# Go Face Recognition API

A Go face recognition API using the `pigo` library for face detection. This internal service processes images from PUBLIC URLs and provides face detection, validation, and visual marking capabilities.

## Features

- **Face Detection**: Detect faces in images from URLs
- **Selfie Validation**: Validate selfie quality based on face count and confidence
- **Visual Detection**: Return images with face markers drawn as circles
- **Health Checks**: Comprehensive health, readiness, and liveness endpoints for Kubernetes
- **Metrics**: Prometheus metrics endpoint for monitoring
- **Graceful Shutdown**: Proper context-based shutdown handling
- **Structured Logging**: JSON-formatted logs with logrus

## Image Optimization (New)

To improve detection performance on large images, the API now performs intelligent preprocessing before running Pigo:

- Automatic resize to an optimal working size (default: longest side 1024 px), preserving aspect ratio
- Adaptive JPEG compression to a target size (default: ~1 MB) using the highest quality that fits
- Alpha images are flattened to opaque white before JPEG encoding

Benefits:
- Faster face detection on large inputs while maintaining detection quality
- Reduced memory and CPU usage
- Consistent throughput under variable traffic

Configuration (environment variables):
- OPT_MAX_SIDE (default 1024): Longest image side cap in pixels before detection
- OPT_TARGET_MAX_BYTES (default 1000000): Target encoded size (bytes) for adaptive compression
- OPT_JPEG_MIN_QUALITY (default 60): Minimum JPEG quality allowed when compressing
- OPT_JPEG_MAX_QUALITY (default 90): Maximum JPEG quality when compressing

Notes:
- Optimization happens after download and decode, and before face detection
- If optimization is not needed (small images), the original is used unchanged
- Metadata in responses indicates when optimization occurred

## Performance Notes

- Resizing large images (e.g., 4000+ px) to ~1024 px typically yields substantial speedups
- Pigo accuracy is retained as faces are still well above the minimum face window size
- Compression uses highest-quality setting under the target size to preserve detail around faces

## API Endpoints

### Face Detection
- `POST /api/v1/detect` - Detect faces in image URL
- `POST /api/v1/validate` - Validate selfie quality
- `POST /api/v1/detect-visual` - Detect faces and return image with circle markers

### Health & Monitoring
- `GET /api/v1/health` - Health check
- `GET /api/v1/ready` - Readiness check
- `GET /api/v1/live` - Liveness check
- `GET /metrics` - Prometheus metrics

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker (optional)

### Local Development

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd face-recognition-api
   go mod download
   ```

2. **Run the application**:
   ```bash
   go run cmd/api/main.go
   ```

3. **Test the API**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/detect \
     -H "Content-Type: application/json" \
     -d '{"image_url": "public_image_uri"}'
   ```

### Docker Deployment

1. **Build image**:
   ```bash
   docker build -t face-recognition-api .
   ```

2. **Run container**:
   ```bash
   docker run -p 8080:8080 face-recognition-api
   ```

## Configuration

The application can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `:8080` | Server port |
| `READ_TIMEOUT` | `30s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `120s` | HTTP idle timeout |
| `MAX_IMAGE_SIZE` | `15728640` | Max image size (15MB) |
| `MAX_WIDTH` | `5000` | Max image width |
| `MAX_HEIGHT` | `5000` | Max image height |
| `PIGO_MIN_SIZE` | `20` | Minimum face size for detection |
| `PIGO_MAX_SIZE` | `1000` | Maximum face size for detection |
| `PIGO_SHIFT_FACTOR` | `0.1` | Shift detection window factor |
| `PIGO_SCALE_FACTOR` | `1.1` | Scale detection window factor |
| `PIGO_IOU_THRESHOLD` | `0.2` | IoU threshold for face clustering |
| `PIGO_MIN_CONFIDENCE` | `5.0` | Minimum confidence threshold |
| `OPT_MAX_SIDE` | `1024` | Longest image side cap (pixels) before detection |
| `OPT_TARGET_MAX_BYTES` | `1000000` | Target encoded size (bytes) for compression |
| `OPT_JPEG_MIN_QUALITY` | `60` | Minimum JPEG quality when compressing |
| `OPT_JPEG_MAX_QUALITY` | `90` | Maximum JPEG quality when compressing |

Note: Pigo defaults are aligned with upstream recommendations per the library README (MaxSize 1000, ScaleFactor 1.1, ShiftFactor 0.1, IoUThreshold 0.2, MinSize 20).

## API Examples

### Face Detection

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/detect \
  -H "Content-Type: application/json" \
  -d '{"image_url": "https://example.com/image.jpg"}'
```

**Response**:
```json
{
  "faces": [
    {
      "x": 150,
      "y": 120,
      "width": 80,
      "height": 80,
      "confidence": 0.95
    }
  ],
  "count": 1,
  "image_metadata": {
    "width": 1024,
    "height": 768,
    "format": "JPEG",
    "size_bytes": 995000,
    "url": "https://example.com/image.jpg",
    "optimized": true,
    "resized": true,
    "original_width": 4032,
    "original_height": 3024,
    "original_size_bytes": 3876543,
    "jpeg_quality_used": 85
  },
  "processing_time_ms": 125.5
}
```

### Visual Detection

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/detect-visual \
  -H "Content-Type: application/json" \
  -d '{
    "image_url": "https://example.com/image.jpg",
    "circle_color": "red",
    "line_width": 3
  }'
```

**Response**:
```json
{
  "image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRgABA...",
  "faces": [...],
  "count": 1,
  "image_metadata": {...},
  "processing_time_ms": 145.8
}
```

### Selfie Validation

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/validate \
  -H "Content-Type: application/json" \
  -d '{
    "image_url": "https://example.com/selfie.jpg",
    "min_faces": 1,
    "max_faces": 1
  }'
```

**Response**:
```json
{
  "is_valid": true,
  "issues": [],
  "confidence": 0.95,
  "face_count": 1
}
```

## Architecture

The application follows a layered architecture:

```
cmd/api/           # Application entry point
internal/
├── handlers/      # HTTP handlers
├── services/      # Business logic
├── models/        # Data structures
├── middleware/    # HTTP middleware
└── config/        # Configuration
```

## Monitoring

- **Health Checks**: Kubernetes-ready health check endpoints:
  - `/api/v1/health` - General health check
  - `/api/v1/ready` - Readiness probe endpoint
  - `/api/v1/live` - Liveness probe endpoint
- **Metrics**: Prometheus metrics available at `/metrics` for cluster monitoring
- **Structured Logging**: JSON-formatted logs with request correlation for centralized logging
- **Performance Tracking**: Processing time metrics for all operations

## Development

### Building

```bash
go build -o bin/face-recognition-api cmd/api/main.go
```

### Docker Build

```bash
docker build -t face-recognition-api:latest .
```

### Code Quality

The project follows Go best practices:
- SOLID principles
- Proper error handling
- Context usage for cancellation
- Structured logging
- Clean architecture patterns

## Dependencies and Credits

This project uses the following open-source libraries. We thank their authors and contributors:

- [esimov/pigo](https://github.com/esimov/pigo) — Pure Go face detection, pupils/eyes localization, facial landmarks
- [golang.org/x/image/draw](https://pkg.go.dev/golang.org/x/image/draw) — High-quality image resampling (CatmullRom)
- [gorilla/mux](https://github.com/gorilla/mux) — HTTP request router and dispatcher
- [sirupsen/logrus](https://github.com/sirupsen/logrus) — Structured logging for Go
- [go-playground/validator/v10](https://github.com/go-playground/validator) — Struct and field validation
- [stretchr/testify](https://github.com/stretchr/testify) — Testing toolkit with assertions
- [prometheus/client_golang](https://github.com/prometheus/client_golang) — Prometheus instrumentation for Go

Additional dependencies may be listed in go.mod and are used under their respective licenses.

## License

This project is licensed under the MIT License.