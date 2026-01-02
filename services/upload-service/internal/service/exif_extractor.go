package service

import (
	"io"
	"log"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"seungpyolee.com/pkg/model"
)

// ExifExtractor handles extraction of photo metadata from image files
type ExifExtractor struct{}

func NewExifExtractor() *ExifExtractor {
	return &ExifExtractor{}
}

// ExtractMetadata reads EXIF data from image file and returns PhotoMetadata
func (e *ExifExtractor) ExtractMetadata(imageData io.Reader) model.PhotoMetadata {
	metadata := model.PhotoMetadata{}

	exifData, err := exif.Decode(imageData)
	if err != nil {
		log.Printf("[EXIF] Could not decode EXIF data: %v", err)
		return metadata
	}

	// Camera Make
	if make, err := exifData.Get(exif.Make); err == nil {
		metadata.CameraMake = sanitizeString(make)
	}

	// Camera Model
	if model, err := exifData.Get(exif.Model); err == nil {
		metadata.CameraModel = sanitizeString(model)
	}

	// Lens Model
	if lens, err := exifData.Get(exif.LensModel); err == nil {
		metadata.LensModel = sanitizeString(lens)
	}

	// Focal Length
	if fl, err := exifData.Get(exif.FocalLength); err == nil {
		metadata.FocalLength = sanitizeRational(fl)
	}

	// F Number (Aperture)
	if fn, err := exifData.Get(exif.FNumber); err == nil {
		metadata.FNumber = sanitizeRational(fn)
	}

	// Exposure Time
	if et, err := exifData.Get(exif.ExposureTime); err == nil {
		metadata.ExposureTime = sanitizeRational(et)
	}

	// ISO
	if iso, err := exifData.Get(exif.ISOSpeedRatings); err == nil {
		metadata.ISO = sanitizeString(iso)
	}

	// DateTime Original (photo capture time)
	if _, err := exifData.Get(exif.DateTimeOriginal); err == nil {
		if dateTime, err := exifData.DateTime(); err == nil {
			metadata.DateTimeOriginal = dateTime
		}
	}

	log.Printf("[EXIF] Successfully extracted metadata: %+v", metadata)
	return metadata
}

// Helper functions to safely extract EXIF values

func sanitizeString(tag *tiff.Tag) string {
	if tag == nil {
		return ""
	}
	if str, err := tag.StringVal(); err == nil {
		return str
	}
	return tag.String()
}

func sanitizeRational(tag *tiff.Tag) string {
	if tag == nil {
		return ""
	}
	rational, err := tag.Rat(0)
	if err != nil {
		return tag.String()
	}
	if rational != nil {
		return rational.String()
	}
	return ""
}
