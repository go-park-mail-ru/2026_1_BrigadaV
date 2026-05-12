package storage

import (
	"context"
	"fmt"
	"io"

	"guidely-app/pkg/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

// NewS3Client создаёт подключение к MinIO/S3 и при необходимости создаёт bucket.
func NewS3Client(cfg *config.Config) (*S3Client, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("BucketExists: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("MakeBucket: %w", err)
		}

		// Делаем bucket публичным для чтения — иначе URL аватара вернёт 403
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, cfg.S3Bucket)

		if err := client.SetBucketPolicy(ctx, cfg.S3Bucket, policy); err != nil {
			// Не фатально — просто файлы не будут публично доступны
			fmt.Printf("warning: SetBucketPolicy failed: %v\n", err)
		}
	}

	scheme := "http"
	if cfg.S3UseSSL {
		scheme = "https"
	}

	return &S3Client{
		client:   client,
		bucket:   cfg.S3Bucket,
		endpoint: fmt.Sprintf("%s://%s", scheme, cfg.S3Endpoint),
		useSSL:   cfg.S3UseSSL,
	}, nil
}

// UploadFile загружает файл в S3 и возвращает публичный URL.
// objectName — путь внутри bucket, например "avatars/uuid.jpg".
// Возвращаемый URL: http://localhost:9000/guidely/avatars/uuid.jpg
func (s *S3Client) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("PutObject: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, objectName)
	return publicURL, nil
}

// DeleteFile удаляет объект из S3 (например при замене аватара).
func (s *S3Client) DeleteFile(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}
