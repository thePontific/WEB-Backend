package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "LAB1/docs" // 👉 Swagger docs (путь должен совпадать с папкой docs после генерации)

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

	// 📘 Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Сбрасываем все удаления при старте
	if err := rep.ResetDeletedStars(); err != nil {
		logrus.Errorf("Ошибка сброса удалённых звезд: %v", err)
	}

	minioService := service.NewMinioService()
	hand := handler.NewHandler(rep, minioService)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
