package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	MinioEndpoint  = "localhost:9000"
	MinioAccessKey = "minio"
	MinioSecretKey = "minio124"
	BucketName     = "cardsandromeda"
	UseSSL         = false
)

type MinioService struct {
	Client *minio.Client
}

func NewMinioService() *MinioService {
	client, err := minio.New(MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(MinioAccessKey, MinioSecretKey, ""),
		Secure: UseSSL,
	})
	if err != nil {
		panic(err)
	}

	// Создаём бакет, если нет
	exists, _ := client.BucketExists(context.Background(), BucketName)
	if !exists {
		client.MakeBucket(context.Background(), BucketName, minio.MakeBucketOptions{})
	}

	return &MinioService{Client: client}
}

// Генерация URL для SPA
func (s *MinioService) GetImageURL(imageName string) string {
	if imageName == "" {
		return ""
	}
	return fmt.Sprintf("http://%s/%s/%s", MinioEndpoint, BucketName, imageName)
}

// Загрузка файла
func (s *MinioService) UploadFile(id int, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Генерация безопасного имени
	fileName := GenerateFileName(id, fileHeader.Filename)

	// Загрузка в Minio
	_, err = s.Client.PutObject(
		context.Background(),
		BucketName,
		fileName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

// Удаление файла
func (s *MinioService) DeleteFile(fileName string) error {
	return s.Client.RemoveObject(context.Background(), BucketName, fileName, minio.RemoveObjectOptions{})
}

// Очистка имени файла: только латиница и цифры
func SanitizeFileName(name string) string {
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			result += string(r)
		}
	}
	return result
}

// Генерация имени на основе ID
func GenerateFileName(id int, original string) string {
	safe := SanitizeFileName(original)
	if strings.Contains(safe, ".") {
		parts := strings.Split(safe, ".")
		ext := parts[len(parts)-1]
		return fmt.Sprintf("star_%d.%s", id, ext)
	}
	return fmt.Sprintf("star_%d", id)
}
