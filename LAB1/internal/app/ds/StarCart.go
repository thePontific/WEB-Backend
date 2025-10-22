// (заявка)
package ds

import (
	"time"

	"github.com/google/uuid"
)

type StarCart struct {
	ID           int        `gorm:"primaryKey;autoIncrement"`
	Status       string     `gorm:"type:varchar(15);not null"`
	DateCreate   time.Time  `gorm:"not null"`
	CreatorID    uuid.UUID  `gorm:"type:uuid;not null"` // Теперь UUID!
	ModeratorID  *uuid.UUID `gorm:"type:uuid"`          // И moderator тоже UUID
	DateFormed   *time.Time
	DateFinished *time.Time
	Comment      string
	Priority     string `gorm:"type:varchar(20)"`

	// Связи
	Creator   User           `gorm:"foreignKey:CreatorID;references:UUID"`
	Moderator *User          `gorm:"foreignKey:ModeratorID;references:UUID"`
	Items     []StarCartItem `gorm:"foreignKey:CartID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Статусы заявок
const (
	StatusDraft     = "черновик"
	StatusDeleted   = "удалён"
	StatusCreated   = "сформирован"
	StatusCompleted = "завершён"
	StatusRejected  = "отклонён"
)
