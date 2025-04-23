package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	config2 "storage_service/internal/config"
	"strings"
)

type S3Client struct {
	client   *s3.Client
	bucket   string
	endpoint string
	url      string
}

func NewS3Client(cfgS3 *config2.S3Config) (*S3Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               cfgS3.Endpoint,
			SigningRegion:     region,
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfgS3.AccessKey,
			cfgS3.SecretKey,
			"",
		)),
		config.WithRegion(cfgS3.Region),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	_, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(cfgS3.Bucket),
	})

	var bne *types.BucketAlreadyExists
	var bno *types.BucketAlreadyOwnedByYou
	if err != nil && !errors.As(err, &bne) && !errors.As(err, &bno) {
		return nil, fmt.Errorf("не удалось создать бакет: %v", err)
	}

	return &S3Client{
		client:   client,
		bucket:   cfgS3.Bucket,
		endpoint: cfgS3.Endpoint,
		url:      cfgS3.URL,
	}, nil
}

func (s *S3Client) UploadImage(ctx context.Context, data []byte, objectKey string) error {
	contentType := http.DetectContentType(data)
	if contentType == "application/octet-stream" {
		ext := filepath.Ext(objectKey)
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	return err
}

func (s *S3Client) DownloadImage(ctx context.Context, objectKey, savePath string) error {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()

	outFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, result.Body)
	return err
}

func (s *S3Client) DeleteImage(ctx context.Context, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	return err
}

func (s *S3Client) ListImages(ctx context.Context, prefix string) ([]string, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	objects := make([]string, 0)
	for _, item := range result.Contents {
		if strings.HasSuffix(*item.Key, ".jpg") ||
			strings.HasSuffix(*item.Key, ".png") ||
			strings.HasSuffix(*item.Key, ".jpeg") {
			objects = append(objects, *item.Key)
		}
	}
	return objects, nil
}

func (s *S3Client) GetImageURL(objectKey string) string {
	return strings.Join([]string{
		s.url,
		s.bucket,
		objectKey,
	}, "/")
}
