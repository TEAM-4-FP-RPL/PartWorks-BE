package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

type ObjectStorage interface {
	Put(ctx context.Context, key string, body io.Reader, contentType string, contentLength int64) (publicURL string, err error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
	KeyFromURL(url string) (key string, ok bool)
}

func NewFromEnv() (ObjectStorage, error) {
	s3Endpoint := strings.TrimSpace(os.Getenv("R2_ENDPOINT"))
	s3Bucket := strings.TrimSpace(os.Getenv("R2_BUCKET"))
	if s3Endpoint == "" || s3Bucket == "" {
		return nil, fmt.Errorf("R2_ENDPOINT and R2_BUCKET must be set")
	}

	accessKeyID := firstNonEmptyEnv("R2_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID")
	secretAccessKey := firstNonEmptyEnv("R2_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY")
	sessionToken := firstNonEmptyEnv("R2_SESSION_TOKEN", "AWS_SESSION_TOKEN")

	if u, err := url.Parse(s3Endpoint); err == nil {
		host := strings.ToLower(strings.TrimSpace(u.Hostname()))
		isAWS := host == "s3.amazonaws.com" || strings.HasSuffix(host, ".amazonaws.com")
		if !isAWS {
			if accessKeyID == "" || secretAccessKey == "" {
				return nil, fmt.Errorf("R2_ACCESS_KEY_ID/R2_SECRET_ACCESS_KEY (or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY) must be set for S3-compatible endpoint")
			}
		}
	}

	region := strings.TrimSpace(os.Getenv("R2_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	publicBaseURL := strings.TrimSpace(os.Getenv("R2_PUBLIC_BASE_URL"))
	if publicBaseURL == "" {
		publicBaseURL = strings.TrimRight(s3Endpoint, "/") + "/" + s3Bucket
	}

	forcePathStyle := parseBoolDefault(os.Getenv("R2_FORCE_PATH_STYLE"), false)

	st, err := NewS3(context.Background(), S3Config{
		Endpoint:        s3Endpoint,
		Region:          region,
		Bucket:          s3Bucket,
		PublicBaseURL:   publicBaseURL,
		ForcePathStyle:  forcePathStyle,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	})
	if err != nil {
		return nil, err
	}
	return st, nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func parseBoolDefault(v string, def bool) bool {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	case "":
		return def
	default:
		return def
	}
}

func sanitizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	clean := path.Clean("/" + key)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || clean == "" || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("invalid key")
	}
	return clean, nil
}
