package repository

import (
	"LAB1/internal/app/ds"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) Register(user *ds.User) error {
	if user.UUID == uuid.Nil {
		user.UUID = uuid.New()
	}
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByUUID(userUUID uuid.UUID) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("uuid = ?", userUUID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	user := &ds.User{}
	if err := r.db.Where("login = ?", login).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) UpdateUserByUUID(userUUID uuid.UUID, updated *ds.User) error {
	var user ds.User
	err := r.db.Where("uuid = ?", userUUID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Обновляем только те поля, которые пришли
	if updated.Login != "" {
		user.Login = updated.Login
	}
	if updated.Password != "" {
		user.Password = updated.Password
	}
	if updated.Role != 0 { // если пришла новая роль
		user.Role = updated.Role
	}

	return r.db.Save(&user).Error
}
func (r *Repository) GetDraftCartByCreatorID(creatorID uuid.UUID) (ds.StarCart, error) {
	var cart ds.StarCart
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&cart).Error
	return cart, err
}
