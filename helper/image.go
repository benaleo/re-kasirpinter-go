package helper

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// IsImageURL checks if the given image string is a URL (starts with http)
func IsImageURL(image string) bool {
	return len(image) >= 4 && image[:4] == "http"
}

// UploadImageToR2 uploads an image to R2 using UUID as filename
// Takes image data (base64), folder name, and R2 service interface
// Returns the public URL and error
type R2Uploader interface {
	UploadFromBase64(ctx context.Context, base64Data, folder, filename string) (string, error)
}

func UploadImageToR2(ctx context.Context, uploader R2Uploader, imageData, folder string) (string, error) {
	if uploader == nil {
		return "", fmt.Errorf("R2 service is not available")
	}

	// Generate UUID for filename
	filename := uuid.New().String()

	// Upload using the R2 service
	return uploader.UploadFromBase64(ctx, imageData, folder, filename)
}
