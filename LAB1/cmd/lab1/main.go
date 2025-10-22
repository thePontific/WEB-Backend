package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "LAB1/docs" // Swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"LAB1/internal/app/config"
	"LAB1/internal/app/dsn"
	"LAB1/internal/app/handler"
	"LAB1/internal/app/redis"
	"LAB1/internal/app/repository"
	"LAB1/internal/pkg"
)

// @title StarCart API
// @version 1.0
// @description Backend для управления заявками и звездами (Лабораторная 4)

// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

func main() {
	ctx := context.Background()
	router := gin.Default()

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 1️⃣ Загружаем конфиг
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// 2️⃣ Подключаем PostgreSQL
	postgresString := dsn.FromEnv()
	fmt.Println("Postgres:", postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	if err := rep.ResetDeletedStars(); err != nil {
		logrus.Errorf("Ошибка сброса удалённых звёзд: %v", err)
	}

	// 3️⃣ Проверяем Redis-подключение
	redisClient, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logrus.Fatalf("Ошибка подключения Redis: %v", err)
	}

	// 4️⃣ Создаём Handler
	h := handler.NewHandler(rep, nil, redisClient, conf.JWTSecret) // nil → MinioService пока не используем

	// 5️⃣ Регистрируем статику и маршруты
	h.RegisterStatic(router)
	h.RegisterRoutes(router)

	// 6️⃣ Запускаем приложение через pkg.App
	app := pkg.NewApp(ctx, conf, router, rep)
	app.RunApp()
}
