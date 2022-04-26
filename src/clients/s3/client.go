package s3

import (
	"context"
	"cutco-camper/src/starketext"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws/session"

	"go.uber.org/zap"

	"time"
)

var appKey = "/app/mockstarket.apk"

const appVersionTag = "version"

type S3Client interface {
	GetLatestAppVersion(context.Context) (string, error)
	// generate a url to download the app
	GenerateAppDownloadURL(context.Context) (string, error)
}

type s3ClientImp struct {
	bucket  *string
	primary *s3.Client
}

func (s *s3ClientImp) GetLatestAppVersion(ctx context.Context) (string, error) {
	logger := starketext.LocalLogger(ctx)
	resp, err := s.primary.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: s.bucket,
		Key:    &appKey,
	})
	if err != nil {
		logger.Errorw("failed to head object", zap.Error(err))
		return "", err
	}
	tag := resp.Metadata[appVersionTag]
	if tag == nil {
		return "", errors.New("failed to load app version")
	}
	return *tag, nil
}

func (s *s3ClientImp) GenerateAppDownloadURL(ctx context.Context) (string, error) {
	return s.presignRequest(ctx, appKey, time.Hour)
}

func (s *s3ClientImp) presignRequest(ctx context.Context, key string, duration time.Duration) (string, error) {
	request, _ := s.primary.GetObjectRequest(&s3.GetObjectInput{
		Bucket: s.bucket,
		Key:    &key,
	})
	presign, err := request.Presign(duration)
	if err != nil {
		starketext.LocalLogger(ctx).Errorw("failed to presign url", zap.Error(err))
		return "", err
	}
	return presign, nil
}

func NewClient(baseSess *session.Session, bucket string) *s3ClientImp {
	c := &s3ClientImp{
		primary: s3.New(baseSess),
	}
	return c
}
