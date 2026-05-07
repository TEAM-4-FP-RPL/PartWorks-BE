package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	PublicBaseURL   string
	ForcePathStyle  bool
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type S3Storage struct {
	bucket        string
	publicBaseURL string
	clients       []*s3.Client
}

func NewS3(_ context.Context, cfg S3Config) (*S3Storage, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	region := strings.TrimSpace(cfg.Region)
	bucket := strings.TrimSpace(cfg.Bucket)
	publicBaseURL := strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	if endpoint == "" || region == "" || bucket == "" || publicBaseURL == "" {
		return nil, fmt.Errorf("invalid s3 config")
	}

	awsCfg := aws.Config{
		Region: region,
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken)
	}

	clients := []*s3.Client{
		newS3Client(awsCfg, endpoint, cfg.ForcePathStyle),
		newS3Client(awsCfg, endpoint, !cfg.ForcePathStyle),
	}

	if leapcellEndpoint, ok := canonicalLeapcellEndpoint(endpoint); ok && leapcellEndpoint != endpoint {
		clients = append(clients, newS3Client(awsCfg, leapcellEndpoint, false))
		clients = append(clients, newS3Client(awsCfg, leapcellEndpoint, true))
	}

	return &S3Storage{
		bucket:        bucket,
		publicBaseURL: publicBaseURL,
		clients:       clients,
	}, nil
}

func (s *S3Storage) Put(ctx context.Context, key string, body io.Reader, contentType string, contentLength int64) (string, error) {
	k, err := sanitizeKey(key)
	if err != nil {
		return "", err
	}

	in := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(k),
		Body:        body,
		ContentType: aws.String(contentType),
	}
	if contentLength > 0 {
		in.ContentLength = aws.Int64(contentLength)
	}

	for i, client := range s.clients {
		if _, err := client.PutObject(ctx, in); err != nil {
			if i == len(s.clients)-1 || !shouldRetryWithAltClient(err) {
				return "", err
			}
			if err := rewindReader(in.Body); err != nil {
				return "", err
			}
			continue
		}

		// Always use presigned URLs since the bucket requires authorization
		presignedURL, err := s.PresignedURL(ctx, k, 24*time.Hour) // 24 hour expiry
		if err != nil {
			return "", err
		}

		return presignedURL, nil
	}
	return "", fmt.Errorf("s3 put failed with no attempts")
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	k, err := sanitizeKey(key)
	if err != nil {
		return err
	}
	in := &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(k)}
	for i, client := range s.clients {
		if _, err = client.DeleteObject(ctx, in); err != nil {
			if i == len(s.clients)-1 || !shouldRetryWithAltClient(err) {
				return err
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("s3 delete failed with no attempts")
}

func (s *S3Storage) PublicURL(key string) string {
	k, err := sanitizeKey(key)
	if err != nil {
		return ""
	}
	return s.publicBaseURL + "/" + k
}

func (s *S3Storage) KeyFromURL(url string) (string, bool) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", false
	}
	prefix := s.publicBaseURL + "/"
	if strings.HasPrefix(url, prefix) {
		k := strings.TrimPrefix(url, prefix)
		k, err := sanitizeKey(k)
		if err != nil {
			return "", false
		}
		return k, true
	}
	return "", false
}

func (s *S3Storage) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	k, err := sanitizeKey(key)
	if err != nil {
		return "", err
	}

	// Try to generate presigned URL with each client until one succeeds
	var presignedURL string
	for i, client := range s.clients {
		presignClient := s3.NewPresignClient(client)

		request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(k),
		}, func(options *s3.PresignOptions) {
			options.Expires = expiry
		})

		if err != nil {
			if i == len(s.clients)-1 || !shouldRetryWithAltClient(err) {
				return "", err
			}
			continue
		}

		presignedURL = request.URL
		break
	}

	return presignedURL, nil
}

func shouldRetryWithAltClient(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	code := strings.ToLower(strings.TrimSpace(apiErr.ErrorCode()))
	switch code {
	case "unauthorized", "accessdenied", "signaturedoesnotmatch", "invalidsignatureexception", "invalidaccesskeyid":
		return true
	default:
		return false
	}
}

func rewindReader(r io.Reader) error {
	seeker, ok := r.(io.Seeker)
	if !ok {
		return fmt.Errorf("unable to retry upload: body is not seekable")
	}
	_, err := seeker.Seek(0, io.SeekStart)
	return err
}

func newS3Client(cfg aws.Config, endpoint string, usePathStyle bool) *s3.Client {
	// Check if this is Cloudflare R2 endpoint
	isCloudflareR2 := strings.Contains(endpoint, ".r2.cloudflarestorage.com")

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = usePathStyle

		// Cloudflare R2 specific configuration
		if isCloudflareR2 {
			// Cloudflare R2 requires path-style addressing
			o.UsePathStyle = true
		}
	})
}

func canonicalLeapcellEndpoint(endpoint string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return "", false
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" || !strings.HasSuffix(host, ".leapcellobj.com") {
		return "", false
	}
	return "https://objstorage.leapcell.io", true
}
