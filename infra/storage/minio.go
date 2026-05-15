package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client    *minio.Client
	bucket    string
	publicURL string // e.g. https://storage-undangan-digital.anggriawan.my.id
}

func NewMinIOClient() (*MinIOClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	publicURL := os.Getenv("MINIO_PUBLIC_URL") // e.g. https://storage-undangan-digital.anggriawan.my.id

	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT, MINIO_ACCESS_KEY, and MINIO_SECRET_KEY are required")
	}
	if bucket == "" {
		bucket = "undangan-uploads"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Always ensure public read policy
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucket)
	_ = client.SetBucketPolicy(ctx, bucket, policy)

	return &MinIOClient{client: client, bucket: bucket, publicURL: publicURL}, nil
}

type PresignResult struct {
	UploadURL string
	PublicURL string
}

// UploadFile uploads multipart file directly and returns public URL
func (m *MinIOClient) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := ""
	for i := len(header.Filename) - 1; i >= 0; i-- {
		if header.Filename[i] == '.' {
			ext = header.Filename[i:]
			break
		}
	}
	objectName := uuid.New().String() + ext
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := m.client.PutObject(ctx, m.bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	publicURL := fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucket, objectName)
	if m.publicURL == "" {
		publicURL = fmt.Sprintf("/%s/%s", m.bucket, objectName)
	}
	return publicURL, nil
}

// PresignPut generates a presigned PUT URL for direct browser upload
// Expires in 15 minutes — enough for any reasonable upload flow
func (m *MinIOClient) PresignPut(ctx context.Context, filename string, contentType string) (*PresignResult, error) {
	ext := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = filename[i:]
			break
		}
	}
	objectName := uuid.New().String() + ext

	reqParams := url.Values{}
	if contentType != "" {
		reqParams.Set("Content-Type", contentType)
	}

	presignedURL, err := m.client.PresignedPutObject(ctx, m.bucket, objectName, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Build public URL for reading the object after upload
	var publicURL string
	if m.publicURL != "" {
		publicURL = fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucket, objectName)
	} else {
		// Fallback: reuse the presigned host but strip auth params
		u := *presignedURL
		u.RawQuery = ""
		publicURL = u.String()
	}

	return &PresignResult{
		UploadURL: presignedURL.String(),
		PublicURL: publicURL,
	}, nil
}
