// (заявка)
package ds

import "time"

type StarCart struct {
	ID           int        `gorm:"primaryKey;not null"`       // ID заявки
	Status       string     `gorm:"type:varchar(15);not null"` // Статус
	DateCreate   time.Time  `gorm:"not null"`                  // Дата создания
	CreatorID    int        `gorm:"not null"`                  // Создатель (пользователь)
	ModeratorID  *int       // Модератор (nullable, пока не назначен)
	DateFormed   *time.Time // Дата формирования заявки (действие создателя)
	DateFinished *time.Time // Дата завершения заявки (действие модератора)

	// Пример поля предметной области
	Comment  string // Любые комментарии
	Priority string `gorm:"type:varchar(20)"` // пример дополнительного поля

	// Элементы заявки
	Items []StarCartItem `gorm:"foreignKey:CartID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Статусы заявок
const (
	StatusDraft     = "черновик"
	StatusDeleted   = "удалён"
	StatusCreated   = "сформирован"
	StatusCompleted = "завершён"
	StatusRejected  = "отклонён"
)
