package ds

// — М-М заявки–услуги
type StarCartItem struct {
	ID       int     `gorm:"primaryKey"`
	CartID   int     `gorm:"not null;uniqueIndex:idx_cart_star"`
	StarID   int     `gorm:"not null;uniqueIndex:idx_cart_star"`
	Quantity int     `gorm:"default:1"`
	Speed    float32 // скорость
	Comment  string

	StarCart *StarCart `gorm:"-"` // <<< игнорируем при миграции
	Star     *Star     `gorm:"foreignKey:StarID"`
}
