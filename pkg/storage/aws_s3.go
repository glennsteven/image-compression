package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"strings"
)

// awsS3 implements the storage interface to provides the ability
// put get delete files to AWS S3.
type awsS3 struct {
	svc *s3.S3
}

// NewAwsS3 creates a new instance AWS S3 Service,
func NewAwsS3(sess *session.Session) Storage {
	return &awsS3{
		svc: s3.New(sess),
	}
}

func (a *awsS3) Put(ctx context.Context, bucket, key string, contents []byte, contentType string) error {
	putInput := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("/storage/%s", key)),
		Body:   strings.NewReader(string(contents)),
		ACL:    aws.String("public-read"),
	}

	if contentType != "" {
		putInput.ContentType = aws.String(contentType)
	}

	if _, err := a.svc.PutObjectWithContext(ctx, &putInput); err != nil {
		return fmt.Errorf("storage create object: %w", err)
	}

	return nil
}

func (a *awsS3) Get(ctx context.Context, bucket, key string) ([]byte, error) {
	o, err := a.svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("/storage/%s", key)),
	})

	if err != nil {
		var aErr awserr.Error
		if errors.As(err, &aErr) &&
			aErr.Code() == s3.ErrCodeNoSuchBucket || aErr.Code() == s3.ErrCodeNoSuchKey {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	defer o.Body.Close()

	b, err := io.ReadAll(o.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return b, nil
}
