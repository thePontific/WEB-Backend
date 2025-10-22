package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "LAB1/docs" // Swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"LAB1/internal/app/config"
	"LAB1/internal/app/dsn"
	"LAB1/internal/app/handler"
	"LAB1/internal/app/repository"
	"LAB1/internal/pkg"
	"LAB1/internal/service"
)

// @title StarCart API
// @version 1.0
// @description Backend для управления заявками и звездами (Лабораторная 4)

// @contact.name API Support
// @contact.url https://example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @schemes http
func main() {
	router := gin.Default()

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Загружаем конфиг
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Строка подключения к PostgreSQL
	postgresString := dsn.FromEnv()
	fmt.Println("Postgres:", postgresString)

	// Репозиторий
	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Сбрасываем все логические удаления при старте
	if err := rep.ResetDeletedStars(); err != nil {
		logrus.Errorf("Ошибка сброса удалённых звезд: %v", err)
	}

	// Сервис для работы с MinIO
	minioService := service.NewMinioService()

	// Создаём handler с секретом JWT из конфигурации
	jwtSecret := conf.JWTSecret // строка из .env или конфигурации
	hand := handler.NewHandler(rep, minioService, jwtSecret)

	// Инициализация приложения
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
