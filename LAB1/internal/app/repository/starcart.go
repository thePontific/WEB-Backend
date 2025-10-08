package repository

import (
	"LAB1/internal/app/ds"
	"time"
)

// ====== Получить заявку по ID ======
func (r *Repository) GetCartByID(cartID int) (ds.Cart, error) {
	var cart ds.Cart
	// Загружаем элементы заявки (CartItem)
	if err := r.db.Preload("Items").First(&cart, cartID).Error; err != nil {
		return ds.Cart{}, err
	}
	return cart, nil
}

// ====== Посчитать количество элементов заявки ======
func (r *Repository) CountCartItems(cartID int) (int, error) {
	var count int64
	if err := r.db.Model(&ds.CartItem{}).Where("cart_id = ?", cartID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// ====== Создать новую заявку ======
func (r *Repository) CreateCart(cart *ds.Cart) error {
	return r.db.Create(cart).Error
}

// ====== Добавить элемент в заявку ======
func (r *Repository) AddCartItem(item *ds.CartItem) error {
	return r.db.Create(item).Error
}

func (r *Repository) RawDeleteCartByID(cartID int) error {
	return r.db.Exec(
		"UPDATE carts SET status = ?, date_finished = ? WHERE id = ?",
		ds.StatusDeleted, time.Now(), cartID,
	).Error
}
