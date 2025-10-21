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
	api.POST("/starcart/add", h.AddStarToStarCart)

	// Группа маршрутов для конкретной корзины
	cart := api.Group("/starcart/:cartID")
	{
		cart.GET("", h.GetStarCartDetails)             // GET /api/starcart/:cartID
		cart.PUT("", h.UpdateStarCartHandler)          // PUT /api/starcart/:cartID
		cart.PUT("/form", h.FormStarCart)              // PUT /api/starcart/:cartID/form
		cart.PUT("/finish", h.FinishStarCart)          // PUT /api/starcart/:cartID/finish
		cart.PUT("/item/:id", h.UpdateStarCartItem)    // PUT /api/starcart/:cartID/item/:id
		cart.DELETE("/item/:id", h.DeleteStarCartItem) // DELETE /api/starcart/:cartID/item/:id
		cart.DELETE("", h.DeleteStarCart)              // DELETE /api/starcart/:cartID
	}

	// ====== Пользователи ======
	api.POST("/users/register", h.RegisterUser)
	api.POST("/users/login", h.LoginUser)
	api.POST("/users/logout", h.LogoutUser)
	api.GET("/users/me", h.GetUser)
	api.PUT("/users/me", h.UpdateUser)
}

// м-к-м пут
// UpdateStarCartItem godoc
// @Summary      Обновить элемент заявки (м-м связь)
// @Description  Изменяет количество, комментарий или скорость для конкретного элемента в заявке
// @Tags         StarCartItem
// @Accept       json
// @Produce      json
// @Param        cartID  path      int  true  "ID заявки"
// @Param        id      path      int  true  "ID элемента в заявке"
// @Param        item    body      ds.StarCartItem  true  "Данные для обновления"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /starcart/{cartID}/item/{id} [put]
func (h *Handler) UpdateStarCartItem(ctx *gin.Context) {
	cartID, err := strconv.Atoi(ctx.Param("cartID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cart ID"})
		return
	}

	itemID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	var input ds.StarCartItem
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование корзины
	cart, err := h.Repository.GetCartByID(cartID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	// Проверяем существование элемента и принадлежность корзине
	item, err := h.Repository.GetStarCartItemByID(itemID)
	if err != nil || item.CartID != cart.ID {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "item not found in this cart"})
		return
	}

	// Обновляем поля
	if input.Quantity != 0 {
		item.Quantity = input.Quantity
	}
	if input.Comment != "" {
		item.Comment = input.Comment
	}
	if input.Speed != 0 {
		item.Speed = float32(input.Speed)
	}

	if err := h.Repository.UpdateStarCartItem(&item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "StarCart item updated"})
}

// ======================
// ==== ЗВЁЗДЫ ====
// ======================

// GetStars godoc
// @Summary Получить список звёзд
// @Description Возвращает список всех звёзд или ищет по названию
// @Tags Stars
// @Accept  json
// @Produce  json
// @Param title query string false "Поиск по названию"
// @Success 200 {array} ds.Star
// @Failure 500 {object} map[string]string
// @Router /api/stars [get]
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

// GetStarDetails godoc
// @Summary Получить подробности о звезде
// @Description Возвращает информацию и ссылку на изображение
// @Tags Stars
// @Produce  json
// @Param id path int true "ID звезды"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/stars/{id} [get]
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

// CreateStar godoc
// @Summary Создать новую звезду
// @Tags Stars
// @Accept  json
// @Produce  json
// @Param star body ds.Star true "Данные звезды"
// @Success 201 {object} ds.Star
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/stars [post]
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

// UpdateStar godoc
// @Summary Обновить данные звезды
// @Tags Stars
// @Accept  json
// @Produce  json
// @Param id path int true "ID звезды"
// @Param star body ds.Star true "Данные звезды"
// @Success 200 {object} ds.Star
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/stars/{id} [put]
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

// DeleteStar godoc
// @Summary Удалить звезду
// @Tags Stars
// @Produce  json
// @Param id path int true "ID звезды"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/stars/{id} [delete]
func (h *Handler) DeleteStar(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	err := h.Repository.DeleteStar(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "star deleted"})
}

// UploadStarImage godoc
// @Summary Загрузить изображение для звезды
// @Tags Stars
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID звезды"
// @Param image formData file true "Файл изображения"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/stars/{id}/image [post]
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

// GetStarCarts godoc
// @Summary Получить список заявок (StarCart)
// @Tags StarCart
// @Produce  json
// @Param from query string false "Дата с"
// @Param to query string false "Дата по"
// @Param status query string false "Статус заявки"
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/starcart [get]
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

// GetStarCartDetails godoc
// @Summary Получить детали заявки
// @Tags StarCart
// @Produce  json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/starcart/{id} [get]
func (h *Handler) GetStarCartDetails(ctx *gin.Context) {
	cartID, err := strconv.Atoi(ctx.Param("cartID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cart ID"})
		return
	}

	cart, err := h.Repository.GetCartByID(cartID)
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

// UpdateStarCartHandler godoc
// @Summary Обновить позиции в заявке
// @Tags StarCart
// @Accept  json
// @Produce  json
// @Param id path int true "ID заявки"
// @Param items body []ds.StarCartItem true "Список позиций"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/starcart/{id} [put]
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

// FormStarCart godoc
// @Summary Сформировать заявку
// @Tags StarCart
// @Produce  json
// @Param id path int true "ID заявки"
// @Success 200 {object} ds.StarCart
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/starcart/{id}/form [put]
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

// FinishStarCart godoc
// @Summary Завершить заявку (одобрить/отклонить)
// @Tags StarCart
// @Produce  json
// @Param id path int true "ID заявки"
// @Param action query string true "complete или reject"
// @Success 200 {object} ds.StarCart
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/starcart/{id}/finish [put]
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

// DeleteStarCart godoc
// @Summary Удалить заявку (логически)
// @Tags StarCart
// @Produce  json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/starcart/{id} [delete]
func (h *Handler) DeleteStarCart(ctx *gin.Context) {
	cartID, err := strconv.Atoi(ctx.Param("cartID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cart ID"})
		return
	}

	cart, err := h.Repository.GetCartByID(cartID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	if cart.CreatorID != CurrentUserID() {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := h.Repository.MarkStarCartAsDeleted(cartID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "starcart logically deleted"})
}

// GetStarCartIcon godoc
// @Summary Получить иконку корзины (draft-заявка)
// @Tags StarCart
// @Produce  json
// @Success 200 {object} map[string]int
// @Router /api/starcart/icon [get]
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

// AddStarToStarCart godoc
// @Summary Добавить звезду в заявку
// @Tags StarCart
// @Accept multipart/form-data
// @Produce  json
// @Param star_id formData int true "ID звезды"
// @Param quantity formData int true "Количество"
// @Param comment formData string false "Комментарий"
// @Success 200 {object} map[string]string
// @Router /api/starcart/add [post]
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

// DeleteStarCartItem godoc
// @Summary Удалить позицию из заявки
// @Tags StarCartItem
// @Produce  json
// @Param id path int true "ID позиции"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/starcart/item/{id} [delete]
func (h *Handler) DeleteStarCartItem(ctx *gin.Context) {
	cartID, err := strconv.Atoi(ctx.Param("cartID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cart ID"})
		return
	}

	itemID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	// Проверяем существование корзины
	cart, err := h.Repository.GetCartByID(cartID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "starcart not found"})
		return
	}

	// Проверяем элемент
	item, err := h.Repository.GetStarCartItemByID(itemID)
	if err != nil || item.CartID != cart.ID {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "item not found in this cart"})
		return
	}

	if err := h.Repository.DeleteStarCartItemByID(itemID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}

// ======================
// ==== ПОЛЬЗОВАТЕЛИ ====
// ======================
// RegisterUser godoc
// @Summary Зарегистрировать нового пользователя
// @Tags Users
// @Accept  json
// @Produce  json
// @Param user body ds.Users true "Данные пользователя"
// @Success 201 {object} ds.Users
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/register [post]
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

// LoginUser godoc
// @Summary Вход пользователя
// @Tags Users
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /api/users/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "login placeholder"})
}

// LogoutUser godoc
// @Summary Выход пользователя
// @Tags Users
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /api/users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "logout placeholder"})
}

// GetUser godoc
// @Summary Получить информацию о текущем пользователе
// @Tags Users
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /api/users/me [get]
func (h *Handler) GetUser(ctx *gin.Context) {
	userID := CurrentUserID()
	user, _ := h.Repository.GetUserByID(userID)
	ctx.JSON(http.StatusOK, gin.H{"login": user.Login, "isModerator": user.IsModerator})
}

// UpdateUser godoc
// @Summary Обновить данные пользователя
// @Tags Users
// @Accept  json
// @Produce  json
// @Param user body ds.Users true "Новые данные"
// @Success 200 {object} ds.Users
// @Failure 400 {object} map[string]string
// @Router /api/users/me [put]
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
