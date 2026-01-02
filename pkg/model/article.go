package model

import "time"

// Photo represents a photo uploaded by a user
type Photo struct {
	PhotoID    string        `json:"photoId" bson:"_id"`
	UserID     string        `json:"userId" bson:"user_id"`     // Reference to User (Foreign Key)
	FileName   string        `json:"fileName" bson:"file_name"` // Original filename
	MimeType   string        `json:"mimeType" bson:"mime_type"` // e.g., "image/jpeg"
	UploadedAt time.Time     `json:"uploadedAt" bson:"uploaded_at"`
	Metadata   PhotoMetadata `json:"metadata" bson:"metadata"` // EXIF and other metadata
}

// PhotoMetadata stores extracted EXIF and technical photo data
type PhotoMetadata struct {
	CameraModel      string    `json:"cameraModel" bson:"camera_model"`
	CameraMake       string    `json:"cameraMake" bson:"camera_make"`
	LensModel        string    `json:"lensModel" bson:"lens_model"`
	FocalLength      string    `json:"focalLength" bson:"focal_length"`   // e.g., "50mm"
	FNumber          string    `json:"fNumber" bson:"f_number"`           // e.g., "f/2.8"
	ExposureTime     string    `json:"exposureTime" bson:"exposure_time"` // e.g., "1/125"
	ISO              string    `json:"iso" bson:"iso"`
	DateTimeOriginal time.Time `json:"dateTimeOriginal" bson:"date_time_original"` // Photo capture time
	Width            int       `json:"width" bson:"width"`
	Height           int       `json:"height" bson:"height"`
}

// PhotoUploadRequest represents the API request for uploading a photo
type PhotoUploadRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// File is handled separately via multipart form data
}

// PhotoQueryResponse represents the API response for photo metadata
type PhotoQueryResponse struct {
	Photo Photo `json:"photo"`
}
