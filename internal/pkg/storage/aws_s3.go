package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"sync"

	"ADRIFT-backend/internal/pkg/validate"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type (
	AwsS3 interface {
		UploadFile(filename string, file *multipart.FileHeader, folderName string, mv ...string) (string, error)
		UpdateFile(objectKey string, f *multipart.FileHeader, mv ...string) (string, error)
		DeleteFile(objectKey string) error
		GetPublicLink(objectKey string) string
		GetObjectKeyFromLink(link string) string
		GetFile(objectKey string) (io.ReadCloser, string, string, error)
		IsOldCloudHostLink(link string) bool
		ConvertOldLinkToObjectKey(link string) string
		Begin() AwsS3
		Commit()
		Rollback()
	}

	action struct {
		actionType string
		key        string
	}

	awsS3 struct {
		client     *s3.Client
		bucket     string
		region     string
		actions    []action
		isRollback bool
	}
)

func NewAwsS3() AwsS3 {
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_KEY"),
			"",
		)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS configuration: %v", err))
	}
	cfg.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	cfg.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired

	client := s3.NewFromConfig(cfg)

	return &awsS3{
		client:     client,
		bucket:     bucket,
		region:     region,
		actions:    nil,
		isRollback: false,
	}
}

func (a *awsS3) UploadFile(filename string, f *multipart.FileHeader, folderName string, mv ...string) (string, error) {
	file, err := f.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	mimetype, err := validate.GetMimeType(file)
	if err != nil {
		return "", err
	}

	if len(mv) > 0 {
		flag := false
		for _, m := range mv {
			if mimetype == m {
				flag = true
				break
			}
		}

		if !flag {
			return "", fmt.Errorf("invalid mimetype")
		}
	}

	objectKey := fmt.Sprintf("%s/%s", folderName, filename)

	_, err = a.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(mimetype),
	})
	if err != nil {
		return "", err
	}

	a.actions = append(a.actions, action{actionType: "upload", key: objectKey})

	return objectKey, nil
}

func (a *awsS3) UpdateFile(objectKey string, f *multipart.FileHeader, mv ...string) (string, error) {
	file, err := f.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	mimetype, err := validate.GetMimeType(file)
	if err != nil {
		return "", err
	}

	if len(mv) > 0 {
		flag := false
		for _, m := range mv {
			if mimetype == m {
				flag = true
				break
			}
		}

		if !flag {
			return "", fmt.Errorf("invalid mimetype")
		}
	}

	_, err = a.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(mimetype),
	})
	if err != nil {
		return "", err
	}
	a.actions = append(a.actions, action{actionType: "update", key: objectKey})

	return objectKey, nil
}

func (a *awsS3) DeleteFile(objectKey string) error {
	_, err := a.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *awsS3) IsOldCloudHostLink(link string) bool {
	return strings.HasPrefix(link, "https://is3.idcloudhost.id/")
}

func (a *awsS3) ConvertOldLinkToObjectKey(link string) string {
	oldPrefix := "https://is3.idcloudhost.id/" + a.bucket + "/"
	if strings.HasPrefix(link, oldPrefix) {
		return strings.TrimPrefix(link, oldPrefix)
	}
	return ""
}

func (a *awsS3) GetPublicLink(objectKey string) string {
	return objectKey
}

func (a *awsS3) GetObjectKeyFromLink(link string) string {
	oldPrefix := "https://is3.cloudhost.id/" + a.bucket + "/"
	if strings.HasPrefix(link, oldPrefix) {
		return strings.TrimPrefix(link, oldPrefix)
	}

	newPrefix := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", a.bucket, a.region)
	if strings.HasPrefix(link, newPrefix) {
		return strings.TrimPrefix(link, newPrefix)
	}

	altPrefix := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/", a.region, a.bucket)
	if strings.HasPrefix(link, altPrefix) {
		return strings.TrimPrefix(link, altPrefix)
	}

	return link
}

func (a *awsS3) GetFile(objectKey string) (io.ReadCloser, string, string, error) {
	result, err := a.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get file from S3: %w", err)
	}

	contentType := "application/pdf"
	if strings.HasSuffix(strings.ToLower(objectKey), ".pdf") {
		contentType = "application/pdf"
	} else if strings.HasSuffix(strings.ToLower(objectKey), ".jpg") ||
		strings.HasSuffix(strings.ToLower(objectKey), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(objectKey), ".png") {
		contentType = "image/png"
	} else if result.ContentType != nil {
		contentType = *result.ContentType
	}

	parts := strings.Split(objectKey, "/")
	filename := parts[len(parts)-1]

	return result.Body, contentType, filename, nil
}

func (a *awsS3) Begin() AwsS3 {
	return &awsS3{
		client:     a.client,
		bucket:     a.bucket,
		region:     a.region,
		actions:    []action{},
		isRollback: true,
	}
}

func (a *awsS3) Commit() {
	a.actions = nil
	a.isRollback = false
}

func (a *awsS3) Rollback() {
	var wg sync.WaitGroup
	errCh := make(chan error, len(a.actions))

	for i := len(a.actions) - 1; i >= 0; i-- {
		action := a.actions[i]
		switch action.actionType {
		case "upload", "update":
			wg.Add(1)
			go func(key string) {
				defer wg.Done()
				if err := a.DeleteFile(key); err != nil {
					errCh <- fmt.Errorf("failed to delete file %s: %v", key, err)
				}
			}(action.key)
		}
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("rollback error: %v\n", err)
		}
	}

	a.Commit()
}
