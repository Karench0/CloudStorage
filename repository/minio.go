package repository

import (
	"CloudStorage/config"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinio() *minio.Client {
	minioClient, err := minio.New(config.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinioAccessKey, config.MinioSecretKey, ""),
		Secure: config.MinioUseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, config.MinioBucket)
	if err != nil {
		log.Printf("Ошибка при проверке bucket: %v", err)
	} else if !exists {
		if err = minioClient.MakeBucket(ctx, config.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			log.Printf("Ошибка при создании bucket: %v", err)
		} else {
			log.Printf("Bucket %s успешно создан", config.MinioBucket)
		}
	}

	return minioClient
}

func BuildObjectKey(userID int, directoryID *int, fileName string) string {
	dirPart := "root"
	if directoryID != nil {
		dirPart = fmt.Sprintf("dir-%d", *directoryID)
	}
	return fmt.Sprintf("users/%d/%s/%s_%s", userID, dirPart, uuid.New().String(), fileName)
}

func PutObject(client *minio.Client, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := client.PutObject(context.Background(), config.MinioBucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func GetObject(client *minio.Client, objectKey string) (*minio.Object, error) {
	return client.GetObject(context.Background(), config.MinioBucket, objectKey, minio.GetObjectOptions{})
}

func DeleteObject(client *minio.Client, objectKey string) error {
	if objectKey == "" {
		return nil
	}
	return client.RemoveObject(context.Background(), config.MinioBucket, objectKey, minio.RemoveObjectOptions{})
}
