package ds

import (
	"LAB1/internal/app/role"

	"github.com/google/uuid"
)

type User struct {
	UUID     uuid.UUID `gorm:"type:uuid;primaryKey"`             // UUID как primary key
	Login    string    `gorm:"type:varchar(50);unique;not null"` // уникальный логин
	Role     role.Role `gorm:"default:0"`                        // 0 = Buyer
	Password string    `gorm:"type:varchar(100);not null"`
}
