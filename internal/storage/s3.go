package storage

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage uploads files to an AWS S3 bucket.
type S3Storage struct {
	client     *s3.Client
	bucketName string
	prefix     string
}

func NewS3Storage(client *s3.Client, bucketName, prefix string) (*S3Storage, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name must be provided")
	}
	return &S3Storage{client: client, bucketName: bucketName, prefix: prefix}, nil
}

func (s *S3Storage) Save(ctx context.Context, relativePath string, r io.Reader) (string, error) {
	key := path.Join(s.prefix, relativePath)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return "", fmt.Errorf("upload to S3: %w", err)
	}
	return fmt.Sprintf("s3://%s/%s", s.bucketName, key), nil
}
