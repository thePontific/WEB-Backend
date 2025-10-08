// (заявка)
package ds

import "time"

// Cart — заявка
type Cart struct {
	ID         int       `gorm:"primaryKey;not null"`       // ID заявки
	Status     string    `gorm:"type:varchar(15);not null"` // Статус
	DateCreate time.Time `gorm:"not null"`                  // Дата создания
	CreatorID  int       `gorm:"not null"`                  // Создатель

	Items []CartItem `gorm:"foreignKey:CartID"` // элементы заявки
}

// Статусы заявок
const (
	StatusDraft     = "черновик"
	StatusDeleted   = "удалён"
	StatusCreated   = "сформирован"
	StatusCompleted = "завершён"
	StatusRejected  = "отклонён"
)
