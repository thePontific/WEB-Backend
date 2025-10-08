package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"LAB1/internal/app/ds"
	"LAB1/internal/app/repository"
	"LAB1/internal/service"
)

type Handler struct {
	Repository   *repository.Repository
	MinioService *service.MinioService
}

// CurrentUserID — фиксированный создатель для лабораторной
func CurrentUserID() int {
	return 1
}

func NewHandler(r *repository.Repository, ms *service.MinioService) *Handler {
	return &Handler{
		Repository:   r,
		MinioService: ms,
	}
}

// ======================
// Статика и маршруты
// ======================
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")

	// ====== Звёзды / услуги ======
	api.GET("/stars", h.GetStars)
	api.GET("/stars/:id", h.GetStarDetails)
	api.POST("/stars", h.CreateStar)
	api.PUT("/stars/:id", h.UpdateStar)
	api.DELETE("/stars/:id", h.DeleteStar)
	api.POST("/stars/:id/image", h.UploadStarImage)

	// ====== Корзина / заявки ======
	api.GET("/cart/icon", h.GetCartIcon)
	api.GET("/cart", h.GetCarts)
	api.GET("/cart/:id", h.GetCartDetails)
	api.POST("/cart/add", h.AddStarToCart)
	api.PUT("/cart/:id", h.UpdateCartHandler)
	api.PUT("/cart/:id/form", h.FormCart)
	api.PUT("/cart/:id/finish", h.FinishCart)
	api.DELETE("/cart/:id", h.DeleteCart)

	// ====== Пользователи ======
	api.POST("/users/register", h.RegisterUser)
	api.POST("/users/login", h.LoginUser)
	api.POST("/users/logout", h.LogoutUser)
	api.GET("/users/me", h.GetUser)
	api.PUT("/users/me", h.UpdateUser)

	api.DELETE("/cart/item/:id", h.DeleteCartItem)

}

// ======================
// ==== ЗВЁЗДЫ ====
// ======================
func (h *Handler) GetStars(ctx *gin.Context) {
	search := ctx.Query("title")
	var stars []ds.Star
	var err error
	if search == "" {
		stars, err = h.Repository.GetStars()
	} else {
		stars, err = h.Repository.SearchStarByTitle(search)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, stars)
}

func (h *Handler) GetStarDetails(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	star, err := h.Repository.GetStar(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "star not found"})
		return
	}
	starURL := h.MinioService.GetImageURL(star.ImageName)
	ctx.JSON(http.StatusOK, gin.H{"star": star, "imageURL": starURL})
}

func (h *Handler) CreateStar(ctx *gin.Context) {
	var input ds.Star
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.Repository.CreateStar(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, input)
}

func (h *Handler) UpdateStar(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var input ds.Star
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	err := h.Repository.UpdateStar(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, input)
}

func (h *Handler) DeleteStar(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	err := h.Repository.DeleteStar(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// TODO: удалить изображение через Minio
	ctx.JSON(http.StatusOK, gin.H{"message": "star deleted"})
}

// Загрузка изображения
func (h *Handler) UploadStarImage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid star ID"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}

	fileName, err := h.MinioService.UploadFile(id, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed upload: " + err.Error()})
		return
	}

	// Обновляем Star в БД
	star, _ := h.Repository.GetStar(id)
	star.ImageName = fileName
	h.Repository.UpdateStar(&star)

	ctx.JSON(http.StatusOK, gin.H{"imageName": fileName})
}

// Получить список корзин с фильтрацией
func (h *Handler) GetCarts(ctx *gin.Context) {
	from := ctx.Query("from") // YYYY-MM-DD
	to := ctx.Query("to")
	status := ctx.Query("status")

	carts, err := h.Repository.GetCartsFiltered(from, to, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, carts)
}

// Получить детали конкретной корзины
func (h *Handler) GetCartDetails(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "cart not found"})
		return
	}

	// Собираем элементы с изображениями
	var items []gin.H
	for _, item := range cart.Items {
		star, _ := h.Repository.GetStar(item.StarID)
		items = append(items, gin.H{
			"star":     star,
			"quantity": item.Quantity,
			"comment":  item.Comment,
			"imageURL": h.MinioService.GetImageURL(star.ImageName),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"cart":  cart,
		"items": items,
	})
}

// Обновление корзины (изменение количества/комментариев)
func (h *Handler) UpdateCartHandler(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var input []ds.StarCartItem
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "cart not found"})
		return
	}

	for _, item := range input {
		h.Repository.UpdateCartItem(&item)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Корзина обновлена"})
}

// Формирование корзины (создатель)
func (h *Handler) FormCart(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "cart not found"})
		return
	}

	if cart.Status != ds.StatusDraft {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cart not draft"})
		return
	}

	now := time.Now()
	cart.Status = ds.StatusCreated
	cart.DateFormed = &now

	if err := h.Repository.UpdateCart(&cart); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// Завершение / отклонение корзины (модератор)
func (h *Handler) FinishCart(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "cart not found"})
		return
	}

	if cart.Status != ds.StatusCreated {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cart not formed"})
		return
	}

	action := ctx.Query("action") // complete / reject
	now := time.Now()

	if action == "complete" {
		cart.Status = ds.StatusCompleted
		cart.DateFinished = &now
	} else if action == "reject" {
		cart.Status = ds.StatusRejected
		cart.DateFinished = &now
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	// TODO: вычисления стоимости, доставки, м-м
	if err := h.Repository.UpdateCart(&cart); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// Удаление корзины (создатель)
func (h *Handler) DeleteCart(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "cart not found"})
		return
	}

	if cart.CreatorID != CurrentUserID() {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	err = h.Repository.RawDeleteCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "cart deleted"})
}

// ======================
// ==== КОРЗИНА / ЗАЯВКИ ====
// ======================
func (h *Handler) GetCartIcon(ctx *gin.Context) {
	userID := CurrentUserID()
	cart, err := h.Repository.GetDraftCartByCreatorID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"cartID": 0, "itemsCount": 0})
		return
	}
	count, _ := h.Repository.CountCartItems(cart.ID)
	ctx.JSON(http.StatusOK, gin.H{"cartID": cart.ID, "itemsCount": count})
}

func (h *Handler) AddStarToCart(ctx *gin.Context) {
	userID := CurrentUserID()
	starID, _ := strconv.Atoi(ctx.PostForm("star_id"))
	qty, _ := strconv.Atoi(ctx.PostForm("quantity"))
	comment := ctx.PostForm("comment")
	if qty < 1 {
		qty = 1
	}
	cart, err := h.Repository.GetDraftCartByCreatorID(userID)
	if err != nil {
		cart = ds.StarCart{CreatorID: userID, Status: ds.StatusDraft, DateCreate: time.Now()}
		h.Repository.CreateCart(&cart)
	}
	item := ds.StarCartItem{CartID: cart.ID, StarID: starID, Quantity: qty, Comment: comment}
	h.Repository.AddCartItem(&item)
	ctx.JSON(http.StatusOK, gin.H{"message": "Звезда добавлена в корзину"})
}

// TODO: остальные методы: GetCarts, GetCartDetails, UpdateCartHandler, FormCart, FinishCart, DeleteCart
// Они будут аналогично реализованы через Repository

// ======================
// ==== ПОЛЬЗОВАТЕЛИ ====
// ======================
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var input ds.Users
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.Repository.CreateUser(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, input)
}

func (h *Handler) LoginUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "login placeholder"})
}

func (h *Handler) LogoutUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "logout placeholder"})
}

func (h *Handler) GetUser(ctx *gin.Context) {
	userID := CurrentUserID()
	user, _ := h.Repository.GetUserByID(userID)
	ctx.JSON(http.StatusOK, gin.H{"login": user.Login, "isModerator": user.IsModerator})
}

func (h *Handler) UpdateUser(ctx *gin.Context) {
	userID := CurrentUserID()
	var input ds.Users
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = userID
	h.Repository.UpdateUser(&input)
	ctx.JSON(http.StatusOK, input)
}

// Удаление элемента из корзины
func (h *Handler) DeleteCartItem(ctx *gin.Context) {
	itemID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	err = h.Repository.DeleteCartItemByID(itemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}
