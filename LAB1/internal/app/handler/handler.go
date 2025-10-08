package handler

// import (
// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"

// 	"LAB1/internal/app/repository"
// 	"LAB1/internal/service"
// )

// type Handler struct {
// 	Repository   *repository.Repository
// 	MinioService *service.MinioService
// }

// func NewHandler(r *repository.Repository, ms *service.MinioService) *Handler {
// 	return &Handler{
// 		Repository:   r,
// 		MinioService: ms,
// 	}
// }

// // RegisterHandler регистрирует маршруты
// func (h *Handler) RegisterHandler(router *gin.Engine) {
// 	router.GET("/Andromeda", h.GetStars)
// 	router.GET("/Andromeda/star/:id", h.GetStarDetails)
// 	router.GET("/Andromeda/starscart/:id", h.GetCartDetails)

// 	// Добавление звезды в корзину
// 	router.POST("/cart/add", h.AddStarToCart)
// 	router.POST("/delete-cart", h.DeleteCart)
// }

// // RegisterStatic регистрирует статику и шаблоны
// func (h *Handler) RegisterStatic(router *gin.Engine) {
// 	router.LoadHTMLGlob("templates/*")
// 	router.Static("/static", "./resources")
// }

// // errorHandler удобный вывод ошибок
// func (h *Handler) errorHandler(ctx *gin.Context, status int, err error) {
// 	logrus.Error(err.Error())
// 	ctx.JSON(status, gin.H{
// 		"status":      "error",
// 		"description": err.Error(),
// 	})
// }
