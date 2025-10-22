package repository

import (
	"LAB1/internal/app/ds"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
	}, nil

}
func (r *Repository) GetStarCartsFiltered(from, to, status string) ([]ds.StarCart, error) {
	var carts []ds.StarCart
	q := r.db.Preload("Items").Where("status != ?", ds.StatusDeleted)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if from != "" {
		q = q.Where("date_formed >= ?", from)
	}
	if to != "" {
		q = q.Where("date_formed <= ?", to)
	}
	err := q.Find(&carts).Error
	return carts, err
}

func (r *Repository) UpdateCart(cart *ds.StarCart) error {
	return r.db.Save(cart).Error
}

// Создание новой услуги (Star)
func (r *Repository) CreateStar(star *ds.Star) error {
	return r.db.Create(star).Error
}

// Обновление услуги
func (r *Repository) UpdateStar(star *ds.Star) error {
	return r.db.Save(star).Error
}

// Создание пользователя
func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

// Обновление пользователя
func (r *Repository) UpdateUser(user *ds.User) error {
	return r.db.Save(user).Error
}

// UpdateCartItem обновляет количество и комментарий элемента корзины
func (r *Repository) UpdateStarCartItem(item *ds.StarCartItem) error {
	return r.db.Model(&ds.StarCartItem{}).
		Where("cart_id = ? AND star_id = ?", item.CartID, item.StarID).
		Updates(map[string]interface{}{
			"quantity": item.Quantity,
			"comment":  item.Comment,
		}).Error
}
func (r *Repository) DeleteStarCartItemByID(id int) error {
	tx := r.db.Exec("DELETE FROM star_cart_items WHERE id = ?", id)
	return tx.Error
}
