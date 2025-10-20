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

	// ====== StarCart / заявки ======
	api.GET("/starcart/icon", h.GetStarCartIcon)
	api.GET("/starcart", h.GetStarCarts)
	api.GET("/starcart/:id", h.GetStarCartDetails)
	api.POST("/starcart/add", h.AddStarToStarCart)
	api.PUT("/starcart/:id", h.UpdateStarCartHandler)
	api.PUT("/starcart/:id/form", h.FormStarCart)
	api.PUT("/starcart/:id/finish", h.FinishStarCart)
	api.DELETE("/starcart/:id", h.DeleteStarCart)

	// ====== Пользователи ======
	api.POST("/users/register", h.RegisterUser)
	api.POST("/users/login", h.LoginUser)
	api.POST("/users/logout", h.LogoutUser)
	api.GET("/users/me", h.GetUser)
	api.PUT("/users/me", h.UpdateUser)

	api.DELETE("/starcart/item/:id", h.DeleteStarCartItem)
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
	ctx.JSON(http.StatusOK, gin.H{"message": "star deleted"})
}

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

	star, _ := h.Repository.GetStar(id)
	star.ImageName = fileName
	h.Repository.UpdateStar(&star)

	ctx.JSON(http.StatusOK, gin.H{"imageName": fileName})
}

// ======================
// ==== STARCART / ЗАЯВКИ ====
// ======================
func (h *Handler) GetStarCarts(ctx *gin.Context) {
	from := ctx.Query("from")
	to := ctx.Query("to")
	status := ctx.Query("status")

	carts, err := h.Repository.GetStarCartsFiltered(from, to, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []gin.H
	for _, c := range carts {
		itemsCount := len(c.Items) // 📌 количество м-м записей
		// 📊 считаем сумму quantity
		var totalQty int
		for _, item := range c.Items {
			totalQty += item.Quantity
		}

		// 📈 считаем среднее значение
		var avgAccuracy float64
		if itemsCount > 0 {
			avgAccuracy = float64(totalQty) / float64(itemsCount)
		}
		response = append(response, gin.H{
			"id":               c.ID,
			"creator_id":       c.CreatorID,
			"status":           c.Status,
			"date_create":      c.DateCreate,
			"star_items_count": itemsCount, // ✅ добавили поле
			"average_quantity": avgAccuracy,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) GetStarCartDetails(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	var items []gin.H
	for _, item := range cart.Items {
		star, _ := h.Repository.GetStar(item.StarID)
		items = append(items, gin.H{
			"star_id":   star.ID,
			"title":     star.Title,
			"quantity":  item.Quantity,
			"speed":     item.Speed,
			"comment":   item.Comment,
			"image_url": h.MinioService.GetImageURL(star.ImageName),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":            cart.ID,
		"status":        cart.Status,
		"date_create":   cart.DateCreate,
		"creator_id":    cart.CreatorID,
		"comment":       cart.Comment,
		"date_formed":   cart.DateFormed,
		"date_finished": cart.DateFinished,
		"items_count":   len(items),
		"items":         items,
	})
}

func (h *Handler) UpdateStarCartHandler(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var input []ds.StarCartItem
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	for _, item := range input {
		h.Repository.UpdateStarCartItem(&item)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "StarCart updated"})
}

func (h *Handler) FormStarCart(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	if cart.Status != ds.StatusDraft {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "starcart not draft"})
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

func (h *Handler) FinishStarCart(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	if cart.Status != ds.StatusCreated {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "starcart not formed"})
		return
	}

	action := ctx.Query("action")
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

	if err := h.Repository.UpdateCart(&cart); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

func (h *Handler) DeleteStarCart(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	cart, err := h.Repository.GetCartByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	if cart.CreatorID != CurrentUserID() {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// ⚡ Вместо удаления — логическое обновление статуса
	err = h.Repository.MarkStarCartAsDeleted(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "starcart logically deleted"})
}

func (h *Handler) GetStarCartIcon(ctx *gin.Context) {
	userID := CurrentUserID()
	cart, err := h.Repository.GetDraftCartByCreatorID(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"starcartID": 0, "itemsCount": 0})
		return
	}
	count, _ := h.Repository.CountCartItems(cart.ID)
	ctx.JSON(http.StatusOK, gin.H{"starcartID": cart.ID, "itemsCount": count})
}

func (h *Handler) AddStarToStarCart(ctx *gin.Context) {
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
	ctx.JSON(http.StatusOK, gin.H{"message": "Star added to StarCart"})
}

func (h *Handler) DeleteStarCartItem(ctx *gin.Context) {
	itemID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	err = h.Repository.DeleteStarCartItemByID(itemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}

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
