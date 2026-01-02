# AI Agent Instructions for Photo Gallery Backend Project

## Project Overview
A **microservices-based photo gallery system** built in Go 1.25.5 with two independent services (Uploader, Gallery) sharing common models and configuration. Uses **MongoDB** for metadata persistence, **Redis** for caching, and **Azure Blob Storage** for photo files. Automatically extracts **EXIF data** from uploaded photos.

## Architecture

### Service Structure
- **Monorepo with Go Workspaces** (`go.work`): Coordinates two services + shared package
- **Shared Layer** (`pkg/`): Models and configuration used across services
  - `pkg/model/article.go`: Photo, PhotoMetadata (EXIF), User data structures
  - `pkg/shared/config.go`: Global constants (cache TTLs, database names)
- **Services** (`services/`):
  - `upload-service`: Write operations
    - `POST /api/upload` → Multipart file upload with EXIF extraction + Azure Blob storage
    - `GET /api/photos` → Retrieve user's photos list
  - `read-service`: Read operations
    - `GET /api/gallery/photo/{photoId}` → Retrieve single photo metadata
    - `GET /api/gallery` → Retrieve user's full photo gallery
    - `GET /api/gallery/filter?startDate=...&endDate=...` → Filter by date range

### Dependency Injection Pattern
Every service uses **layered architecture** with explicit interface contracts:
1. **Handler** (HTTP Transport) → depends on Service interface
2. **Service** (Business Logic) → depends on Repository interfaces
3. **Repository** (Infrastructure) → MongoDB + Azure Blob + Redis implementations

**Example from upload-service/cmd/main.go**:
```go
cosmosRepo := repository.NewCosmosDBRepository(...)  // MongoDB for metadata
blobRepo := repository.NewAzureBlobRepository(...)   // Azure Blob for files
uploaderSvc := service.NewUploaderService(cosmosRepo, blobRepo, redisRepo)
uploaderHandler := handler.NewUploaderHandler(uploaderSvc)
```

## Data Models & Storage Strategy

### Core Entities
- **Photo** (`PhotoID`, `UserID`, `BlobURL`, `UploadedAt`, `Metadata`)
  - Stores reference to Azure Blob file, NOT the actual file
  - `BlobURL` points to Azure Blob Storage URL for direct download
  - Multi-user: scoped by `UserID` foreign key
  - Uses `bson` tags for MongoDB serialization
- **PhotoMetadata** (EXIF data extraction)
  - `CameraModel`, `LensModel`, `FocalLength`, `FNumber`, `ExposureTime`, `ISO`
  - `DateTimeOriginal`: Photo capture time (different from upload time)
  - `Width`, `Height`: Image dimensions
  - Extracted automatically via `goexif` library during upload
- **User** (`UserID`, `Email`, `Name`, `CreatedAt`)
  - Basic user record for multi-user support

### Critical Pattern: BSON Serialization
All models use both `json` and `bson` tags:
```go
type Photo struct {
    PhotoID   string `json:"photoId" bson:"_id"`
    UserID    string `json:"userId" bson:"user_id"`
    BlobURL   string `json:"blobUrl" bson:"blob_url"`
    Metadata  PhotoMetadata `json:"metadata" bson:"metadata"`
}
```

## Cloud Storage: Azure Blob Integration

### Upload Flow (Upload Service)
1. **Receive multipart form**: `file`, `title`, `description`
2. **Generate blob name**: `{userID}/{photoID}` (hierarchical path)
3. **Upload to Azure**: `AzureBlobRepository.UploadBlob()` → returns public URL
4. **Extract EXIF**: `ExifExtractor.ExtractMetadata()` from file stream
5. **Save metadata**: MongoDB document with BlobURL + EXIF
6. **Cache**: Redis stores photo metadata for 30 minutes

### Environment Variables (Azure)
- `AZURE_STORAGE_CONNECTION_STRING`: Required for blob uploads
  - Example: `DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;`
- `AZURE_STORAGE_CONTAINER_NAME`: Defaults to `photos` if not set

### Blob URL Pattern
After upload, client accesses photos via:
```
https://{accountname}.blob.core.windows.net/photos/{userID}/{photoID}
```

## Multi-User Authorization Pattern

### Header-Based User Identification
All endpoints require `X-User-ID` header:
```bash
curl -H "X-User-ID: user123" GET /api/gallery
```

**Current Implementation**: Simple header-based (for MVP)
**Production Path**: Replace with JWT token validation

### Data Isolation
- Uploader: Stores photos only for authenticated user
- Gallery: Returns only user's own photos
- Handler enforces: `if photo.UserID != userID { forbidden }`

## EXIF Data Extraction

### Supported Metadata
Library: `github.com/rwcarlsen/goexif`

Extracted fields:
- Camera Make & Model
- Lens Model
- Focal Length (e.g., "50mm")
- F-Number (e.g., "f/2.8")
- Exposure Time (e.g., "1/125")
- ISO
- Date/Time Original (photo capture timestamp)
- Image dimensions (Width × Height)

### Extraction Timing
- Automatic during upload in `ExifExtractor.ExtractMetadata()`
- Failures are logged but don't block upload
- Missing EXIF fields return empty strings (no validation error)

## Caching Strategy

### Redis Usage
- **TTL**: 30 minutes for photo metadata (`PhotoCacheTTL`)
- **Invalidation**: Uploader invalidates cache after new upload
- **Stampede Prevention**: Gallery uses `golang.org/x/sync/singleflight` to prevent concurrent DB hits

### Gallery Read Flow
1. Check Redis cache first → Return if hit
2. Singleflight lock prevents multiple goroutines querying DB simultaneously
3. On cache miss: Query MongoDB → Async update Redis in background
4. Cache failures are non-fatal (service remains functional without cache)

## Environment Configuration

### Required Environment Variables

**Uploader Service**:
- `COSMOS_URI`: MongoDB connection (defaults to `mongodb://localhost:27017`)
- `REDIS_ADDR`: Redis server (defaults to `localhost:6379`)
- `AZURE_STORAGE_CONNECTION_STRING`: **REQUIRED** - Azure Blob connection
- `AZURE_STORAGE_CONTAINER_NAME`: Blob container name (defaults to `photos`)

**Gallery Service**:
- `COSMOS_URI`: MongoDB connection
- `REDIS_ADDR`: Redis server

**Note**: Services expect local defaults for development; Docker Compose in root provides full stack with Azure setup template.

## Build & Run

### Local Development
```bash
# Build both services using go.work
go build ./services/editor-service/cmd/
go build ./services/reader-service/cmd/

# Run services (both open different ports if configured)
./editor-service/cmd/editor-service    # :8080 (Uploader)
./reader-service/cmd/reader-service    # :8081 (Gallery)
```

### Docker
```bash
# Set Azure credentials before running
export AZURE_STORAGE_CONNECTION_STRING="..."

docker-compose up  # Starts MongoDB, Redis, both services
```

## API Endpoints & Testing

### Uploader Service (Port 8080)

**Upload photo**:
```bash
POST /api/upload
Headers: X-User-ID: user123, Content-Type: multipart/form-data
Form data: file (binary), title (string), description (string)

Response: {"photoId": "abc-123", "message": "Photo uploaded successfully"}
```

**Get user's photos**:
```bash
GET /api/photos
Headers: X-User-ID: user123

Response: {"photos": [...], "count": 5}
```

### Gallery Service (Port 8081)

**Get single photo**:
```bash
GET /api/gallery/photo/{photoId}
Headers: X-User-ID: user123

Response: {Photo with metadata and EXIF data}
```

**Get full gallery**:
```bash
GET /api/gallery
Headers: X-User-ID: user123

Response: {"photos": [...], "count": 5}
```

**Filter by date**:
```bash
GET /api/gallery/filter?startDate=2024-01-01T00:00:00Z&endDate=2024-12-31T23:59:59Z
Headers: X-User-ID: user123

Response: {"photos": [...], "count": 3}
```

**Health checks**:
```bash
GET /health  # Both services
```

## Key Developer Conventions

1. **File Storage**: Photos stored in Azure Blob, NOT in MongoDB → BlobURL field is reference, not binary data
2. **User Scoping**: All queries filter by `user_id` → Multi-user isolation is enforced at repository layer
3. **EXIF Extraction**: Automatic but non-blocking → Failures logged, upload succeeds
4. **Errors Don't Stop Workflows**: Failed cache operations log warnings but continue
5. **Async Cache Updates**: Gallery updates cache asynchronously to avoid blocking responses
6. **Repository Interfaces**: Define clear contracts before implementing
7. **Service as Coordinator**: Service layer orchestrates between multiple repositories (DB, Blob, Cache)
8. **Header-Based Auth**: `X-User-ID` for MVP; replace with JWT in production

## Important Files to Reference

- [../pkg/model/article.go](../pkg/model/article.go) — Photo, PhotoMetadata, User data contracts
- [../pkg/shared/config.go](../pkg/shared/config.go) — Cache TTLs, database names, container settings
- [../services/editor-service/internal/repository/azureblob_repo.go](../services/upload-service/internal/repository/azureblob_repo.go) — Azure Blob upload/delete
- [../services/editor-service/internal/service/exif_extractor.go](../services/upload-service/internal/service/exif_extractor.go) — EXIF extraction logic
- [../services/editor-service/internal/service/editor_service.go](../services/upload-service/internal/service/editor_service.go) — Upload orchestration
- [../services/reader-service/internal/service/reader_service.go](../services/read-service/internal/service/reader_service.go) — Singleflight + caching pattern
- [../docker-compose.yml](../docker-compose.yml) — Service setup with Azure template
- [copilot-instructions.md](copilot-instructions.md) — This file

## Known Patterns & Limitations

- **No JWT yet**: Using simple header-based auth for MVP
- **No photo deletion**: API doesn't implement delete (can add to uploader)
- **No shared HTTP client**: Each service makes own Azure/MongoDB calls
- **No centralized logging**: Uses standard `log` package
- **No message queue**: Direct DB + cache synchronization only
- **Synchronous EXIF extraction**: Done during upload (could be async for large files)
