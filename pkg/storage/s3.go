package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	region        string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(config S3Config) (*S3Storage, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	return &S3Storage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        config.Bucket,
		region:        config.Region,
	}, nil
}

func (s3s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, info FileInfo) error {
	_, err := s3s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(info.ContentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	return nil
}

func (s3s *S3Storage) GetDownloadURL(ctx context.Context, key string) (string, error) {
	// Use presigned URL for better security
	presignResult, err := s3s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

func (s3s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s3s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

func (s3s *S3Storage) IsAvailable(ctx context.Context) bool {
	// Try to list objects to check if S3 is accessible
	_, err := s3s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s3s.bucket),
		MaxKeys: aws.Int32(1),
	})
	return err == nil
}
