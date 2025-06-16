package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
)

// S3UploadHandler handles file uploads to AWS S3
func S3UploadHandler(c *gin.Context) {
	userNumber, exists := c.Get("user_number")
	if !exists || userNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_number is required in context (check authentication)"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	// Folder structure: users/{userNumber}/filename
	s3Key := fmt.Sprintf("users/%s/%d_%s", userNumber, time.Now().UnixNano(), header.Filename)

	s3URL, err := uploadToS3(file, header, s3Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": s3URL})
}

func uploadToS3(file multipart.File, header *multipart.FileHeader, key string) (string, error) {
	awsRegion := os.Getenv("AWS_REGION")
	awsBucket := os.Getenv("AWS_BUCKET")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})
	if err != nil {
		return "", err
	}
	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(awsBucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", err
	}
	return result.Location, nil
}
