package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type R2Service struct {
	Client     *s3.Client
	BucketName string
	PublicURL  string
}

func NewR2Service() (*R2Service, error) {
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")
	region := os.Getenv("R2_REGION")
	publicURL := os.Getenv("CLOUD_STORAGE_URL")

	if accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing R2 credentials or bucket name")
	}

	// Create AWS configuration with R2 credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
				}, nil
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// R2 uses S3-compatible API, but we need to set the custom endpoint
		// Extract account ID from the public URL to construct the endpoint
		if strings.Contains(publicURL, "r2.cloudflarestorage.com") {
			// Parse the URL to get the account ID
			parts := strings.Split(publicURL, ".")
			if len(parts) > 0 {
				accountID := strings.TrimPrefix(parts[0], "https://")
				o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
			}
		}
	})

	return &R2Service{
		Client:     client,
		BucketName: bucketName,
		PublicURL:  publicURL,
	}, nil
}

// UploadFromBase64 uploads a file to R2 from base64 encoded string
func (r *R2Service) UploadFromBase64(ctx context.Context, base64Data, folder, filename string) (string, error) {
	// Decode base64
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64Data))

	// Detect content type
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	data := buf.Bytes()

	// Try to detect if it's an image and get the format
	contentType := http.DetectContentType(data)
	var ext string

	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	default:
		// If we can't detect, try to decode as image
		_, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			ext = "." + format
		} else {
			ext = ".bin"
		}
	}

	// If filename doesn't have extension, add detected extension
	if !strings.HasSuffix(filename, ext) {
		filename = filename + ext
	}

	// Create the key path
	key := fmt.Sprintf("%s/%s", folder, filename)

	// Upload to R2
	_, err := r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(r.PublicURL, "/"), key)
	return publicURL, nil
}

// UploadFile uploads a file from multipart form to R2
func (r *R2Service) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	defer file.Close()

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	// Create the key path
	ext := filepath.Ext(header.Filename)
	key := fmt.Sprintf("%s/%s%s", folder, header.Filename, ext)

	// Upload to R2
	_, err = r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(r.PublicURL, "/"), key)
	return publicURL, nil
}

// DeleteFile deletes a file from R2
func (r *R2Service) DeleteFile(ctx context.Context, key string) error {
	_, err := r.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}
	return nil
}

// DeleteFileByURL deletes a file from R2 using its public URL
func (r *R2Service) DeleteFileByURL(ctx context.Context, publicURL string) error {
	// Extract key from public URL
	key := strings.TrimPrefix(publicURL, r.PublicURL+"/")
	if key == publicURL {
		// Try alternative parsing
		key = strings.TrimPrefix(publicURL, strings.TrimSuffix(r.PublicURL, "/")+"/")
	}

	if key == "" {
		return fmt.Errorf("failed to extract key from URL")
	}

	return r.DeleteFile(ctx, key)
}
