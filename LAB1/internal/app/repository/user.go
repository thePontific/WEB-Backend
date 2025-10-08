package repository

import "LAB1/internal/app/ds"

// ====== Получить пользователя по ID ======
func (r *Repository) GetUserByID(userID int) (ds.Users, error) {
	var user ds.Users
	if err := r.db.First(&user, userID).Error; err != nil {
		return ds.Users{}, err
	}
	return user, nil
}
func (r *Repository) GetDraftCartByCreatorID(creatorID int) (ds.StarCart, error) {
	var cart ds.StarCart
	err := r.db.Preload("Items").
		Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).
		First(&cart).Error
	if err != nil {
		return ds.StarCart{}, err
	}
	return cart, nil
}
