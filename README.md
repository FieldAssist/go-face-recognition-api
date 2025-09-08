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
| `PIGO_MIN_SIZE` | `25` | Minimum face size for detection |
| `PIGO_MAX_SIZE` | `1000` | Maximum face size for detection |
| `PIGO_MIN_CONFIDENCE` | `12.0` | Minimum confidence threshold |
| `PIGO_IOU_THRESHOLD` | `0.6` | IoU threshold for face clustering |

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
    "width": 640,
    "height": 480,
    "format": "JPEG",
    "size_bytes": 245760,
    "url": "https://example.com/image.jpg"
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

## License

This project is licensed under the MIT License.