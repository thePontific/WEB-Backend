package repository

import (
	"LAB1/internal/app/ds"
	"time"
)

// ====== Получить заявку по ID ======
func (r *Repository) GetCartByID(cartID int) (ds.StarCart, error) {
	var cart ds.StarCart
	// Загружаем элементы заявки (CartItem)
	if err := r.db.Preload("Items").First(&cart, cartID).Error; err != nil {
		return ds.StarCart{}, err
	}
	return cart, nil
}

// ====== Посчитать количество элементов заявки ======
func (r *Repository) CountCartItems(cartID int) (int, error) {
	var count int64
	if err := r.db.Model(&ds.StarCartItem{}).Where("cart_id = ?", cartID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// ====== Создать новую заявку ======
func (r *Repository) CreateCart(cart *ds.StarCart) error {
	return r.db.Create(cart).Error
}

// ====== Добавить элемент в заявку ======
func (r *Repository) AddCartItem(item *ds.StarCartItem) error {
	return r.db.Create(item).Error
}

func (r *Repository) RawDeleteCartByID(cartID int) error {
	return r.db.Exec(
		"UPDATE star_carts SET status = ?, date_finished = ? WHERE id = ?",
		ds.StatusDeleted, time.Now(), cartID,
	).Error
}
func (r *Repository) MarkStarCartAsDeleted(id int) error {
	query := `UPDATE starcarts SET status = ? WHERE id = ?`
	result := r.db.Exec(query, "удалён", id)
	return result.Error
}
