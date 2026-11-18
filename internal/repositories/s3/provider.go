package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	appconfig "github.com/IA-Generative/mozilla-extension-gate/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
)

type S3Provider struct {
	client *s3.Client
	bucket string
	logger *zap.Logger
}

func NewS3Provider(config appconfig.S3Config, logger *zap.Logger) (*S3Provider, error) {
	if config.Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	if config.Region == "" {
		config.Region = "us-east-1"
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return nil, fmt.Errorf("S3 credentials are required")
	}

	ctx := context.Background()
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(config.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				config.AccessKeyID,
				config.SecretAccessKey,
				"",
			),
		),
	}
	if config.TLSInsecureSkipVerify {
		opts = append(opts, awsconfig.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading S3 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if config.Endpoint != "" {
			o.BaseEndpoint = aws.String(config.Endpoint)
			o.UsePathStyle = config.UsePathStyle
		}
	})
	if config.CreateBucketIfNotExists {
		// Create the bucket if not exists
		_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: &config.Bucket,
		})
		if err != nil {
			_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: &config.Bucket,
			})
			if err != nil {
				return nil, fmt.Errorf("creating S3 bucket: %w", err)
			}
		}
	}
	return &S3Provider{
		client: client,
		bucket: config.Bucket,
		logger: logger,
	}, nil
}

func (p *S3Provider) Upload(ctx context.Context, key string, body []byte) error {
	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
		Body:   bytes.NewReader(body),
	})
	return err
}

func (p *S3Provider) Download(ctx context.Context, key string) ([]byte, error) {
	output, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(output.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *S3Provider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})
	return err
}

func (p *S3Provider) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(p.client, &s3.ListObjectsV2Input{
		Bucket: &p.bucket,
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			key := path.Base(*obj.Key)
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (p *S3Provider) GetFile(ctx context.Context, key string) (*io.ReadCloser, error) {
	output, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	return &output.Body, nil
}

func (p *S3Provider) GetFileLength(ctx context.Context, key string) (*int64, error) {
	output, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:       &p.bucket,
		Key:          &key,
		ChecksumMode: "ENABLED",
	})
	if err != nil {
		return nil, err
	}
	return output.ContentLength, nil
}

func (p *S3Provider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})
	if err != nil {
		if _, ok := errors.AsType[*s3types.NotFound](err); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
