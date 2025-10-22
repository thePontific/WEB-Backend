package pkg

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"LAB1/internal/app/config"
	"LAB1/internal/app/handler"
	"LAB1/internal/app/redis"
	"LAB1/internal/app/repository"
	"LAB1/internal/service"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(ctx context.Context, c *config.Config, r *gin.Engine, repo *repository.Repository) *Application {
	// Инициализируем MinIO
	minioService := service.NewMinioService()

	// ✅ Инициализируем Redis (обрати внимание на cfg.Redis)
	redisClient, err := redis.New(ctx, c.Redis)
	if err != nil {
		logrus.Fatalf("failed to connect to Redis: %v", err)
	}

	// Создаём Handler
	h := handler.NewHandler(repo, minioService, redisClient, c.JWTSecret)

	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
