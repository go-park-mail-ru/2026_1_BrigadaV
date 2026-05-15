package storage

import (
	"context"
	"fmt"
	"io"
	"log"

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

func NewS3Client(cfg *config.Config) (*S3Client, error) {
	if !cfg.S3Enabled {
		log.Println("S3 is disabled via S3_ENABLED=false")
		return nil, nil
	}

	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	})
	if err != nil {
		log.Printf("WARNING: minio.New failed: %v; S3 features will be unavailable", err)
		return nil, nil
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		log.Printf("WARNING: BucketExists failed: %v; S3 features will be unavailable", err)
		return nil, nil
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{}); err != nil {
			log.Printf("WARNING: MakeBucket failed: %v; S3 features will be unavailable", err)
			return nil, nil
		}

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
			log.Printf("WARNING: SetBucketPolicy failed: %v (files will not be public)", err)
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

func (s *S3Client) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("S3 client is not initialized")
	}
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("PutObject: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, objectName)
	return publicURL, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, objectName string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("S3 client is not initialized")
	}
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}
