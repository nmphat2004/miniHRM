package avatar

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Store writes private avatar objects and creates short-lived read URLs.
type Store interface {
	Put(context.Context, string, string, []byte) (string, error)
	Delete(context.Context, string) error
	URL(context.Context, string) (string, error)
}

type S3Store struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

func NewS3Store(ctx context.Context, region, bucket string) (*S3Store, error) {
	if strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}
	if strings.TrimSpace(region) == "" {
		region = "ap-southeast-1"
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load AWS config for avatar S3: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return &S3Store{client: client, presigner: s3.NewPresignClient(client), bucket: bucket}, nil
}

func (s *S3Store) Put(ctx context.Context, employeeID, contentType string, data []byte) (string, error) {
	if employeeID == "" || len(data) == 0 {
		return "", fmt.Errorf("employee ID and avatar data are required")
	}
	extension := ".jpg"
	if contentType == "image/png" {
		extension = ".png"
	}
	key := fmt.Sprintf("employees/%s/avatar/%s%s", employeeID, uuid.NewString(), extension)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(key), Body: bytes.NewReader(data),
		ContentType: aws.String(contentType), ServerSideEncryption: "AES256",
	})
	if err != nil {
		return "", fmt.Errorf("upload avatar to S3: %w", err)
	}
	return key, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete avatar from S3: %w", err)
	}
	return nil
}

func (s *S3Store) URL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	result, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return "", fmt.Errorf("create avatar URL: %w", err)
	}
	return result.URL, nil
}
