package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MinioClient struct {
	Client     *s3.Client
	BucketName string
	Endpoint   string
	UseSSL     bool
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func NewMinioClient() (*MinioClient, error) {
	cfg := Config{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		Bucket:    os.Getenv("MINIO_BUCKET_NAME"),
		UseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
	}

	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("missing MinIO environment variables")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &MinioClient{
		Client:     client,
		BucketName: cfg.Bucket,
		Endpoint:   cfg.Endpoint,
		UseSSL:     cfg.UseSSL,
	}, nil
}

func (m *MinioClient) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	_, err := m.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(m.BucketName),
		Key:         aws.String(name),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	var url string
	if m.UseSSL {
		url = fmt.Sprintf("https://%s/%s/%s", strings.TrimPrefix(m.Endpoint, "https://"), m.BucketName, name)
	} else {
		url = fmt.Sprintf("http://%s/%s/%s", strings.TrimPrefix(m.Endpoint, "http://"), m.BucketName, name)
	}

	return url, nil
}

func (m *MinioClient) DeleteFile(ctx context.Context, fileURL string) error {
	parts := strings.Split(fileURL, "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid file URL")
	}
	fileName := parts[len(parts)-1]

	_, err := m.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(m.BucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (m *MinioClient) GetFileURL(fileName string) string {
	var url string
	if m.UseSSL {
		url = fmt.Sprintf("https://%s/%s/%s", strings.TrimPrefix(m.Endpoint, "https://"), m.BucketName, fileName)
	} else {
		url = fmt.Sprintf("http://%s/%s/%s", strings.TrimPrefix(m.Endpoint, "http://"), m.BucketName, fileName)
	}
	return url
}
