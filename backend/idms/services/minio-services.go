package services

import (
	"context"
	"fmt"
	// "net/url"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"time"
)

// MinIOService wraps the MinIO client
type MinIOService struct {
	client     *minio.Client
	tempBucket string
	mainBucket string
}

// NewMinIOService creates a new MinIO service
func NewMinIOService(endpoint, accessKey, secretKey, tempBucket, mainBucket string, useSSL bool) (*MinIOService, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	service := &MinIOService{
		client:     minioClient,
		tempBucket: tempBucket,
		mainBucket: mainBucket,
	}

	// Ensure buckets exist
	ctx := context.Background()
	if err := service.ensureBucketExists(ctx, tempBucket); err != nil {
		return nil, fmt.Errorf("failed to create temp bucket: %w", err)
	}
	if err := service.ensureBucketExists(ctx, mainBucket); err != nil {
		return nil, fmt.Errorf("failed to create main bucket: %w", err)
	}

	return service, nil
}

// ensureBucketExists creates bucket if it doesn't exist
func (s *MinIOService) ensureBucketExists(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		fmt.Printf("Bucket '%s' created successfully\n", bucketName)
	}

	return nil
}

// GenerateUploadURL generates a presigned URL for uploading files
func (s *MinIOService) GenerateUploadURL(key string) (string, error) {

	// Set request parameters for PUT object
	presignedURL, err := s.client.PresignedPutObject(
		context.Background(),
		s.tempBucket,
		key,
		5*time.Minute, // 15 minutes expiration
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return presignedURL.String(), nil
}

// GenerateDownloadURL generates a presigned URL for downloading files
func (s *MinIOService) GenerateDownloadURL(key string, bucket string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = s.mainBucket
	}

	if expiry == 0 {
		expiry = 1 * time.Hour // Default 1 hour
	}

	presignedURL, err := s.client.PresignedGetObject(
		context.Background(),
		bucket,
		key,
		expiry,
		nil, // no additional request parameters
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return presignedURL.String(), nil
}

// MoveFile moves a file from temp bucket to main bucket
func (s *MinIOService) MoveFile(ctx context.Context, key string) error {
	// Copy object from temp to main bucket
	src := minio.CopySrcOptions{
		Bucket: s.tempBucket,
		Object: key,
	}
	dst := minio.CopyDestOptions{
		Bucket: s.mainBucket,
		Object: key,
	}

	_, err := s.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Remove object from temp bucket
	err = s.client.RemoveObject(ctx, s.tempBucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove temp file: %w", err)
	}

	return nil
}

// ListFiles lists files in a bucket
func (s *MinIOService) ListFiles(bucket string, prefix string) ([]minio.ObjectInfo, error) {
	if bucket == "" {
		bucket = s.mainBucket
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var objects []minio.ObjectInfo
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// DeleteFile deletes a file from bucket
func (s *MinIOService) DeleteFile(ctx context.Context, bucket, key string) error {
	if bucket == "" {
		bucket = s.mainBucket
	}

	err := s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileInfo gets information about a file
func (s *MinIOService) GetFileInfo(bucket, key string) (minio.ObjectInfo, error) {
	if bucket == "" {
		bucket = s.mainBucket
	}

	objInfo, err := s.client.StatObject(context.Background(), bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}

	return objInfo, nil
}

// custome function to download the minio files

func (s *MinIOService) DownloadFile(key string) (string, error) {
	localpath:=fmt.Sprintf("/tmp/%s",key)
	var err error
	err = s.client.FGetObject(
		context.Background(),
		s.tempBucket,
		key,
		localpath,
		minio.GetObjectOptions{})

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return localpath,nil
}

// create the meta data of the file in the mainbucket(when its scanned)
func(s * MinIOService) MakeMetaData(key string) error{

	

	return nil
}