package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "github.com/oscar/oscar/internal/config"
)

type R2Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Client(cfg *appconfig.R2Config) (*R2Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID),
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
		o.UsePathStyle = false
	})

	return &R2Client{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

func (r *R2Client) EnsureBucket(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucket),
	})
	if err != nil {
		_, err = r.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(r.bucket),
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return nil
}

func (r *R2Client) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(objectName),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

func (r *R2Client) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	resp, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return resp.Body, nil
}

func (r *R2Client) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.client)
	resp, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}
	return resp.URL, nil
}

func (r *R2Client) GetPresignedPutURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.client)
	resp, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to get presigned put URL: %w", err)
	}
	return resp.URL, nil
}

func (r *R2Client) Delete(ctx context.Context, objectName string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (r *R2Client) List(ctx context.Context, prefix string) ([]string, error) {
	var objects []string
	paginator := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			if obj.Key != nil {
				objects = append(objects, *obj.Key)
			}
		}
	}
	return objects, nil
}

func (r *R2Client) GetPublicURL(objectName string) string {
	if r.publicURL != "" {
		return r.publicURL + "/" + objectName
	}
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", r.bucket, objectName)
}
