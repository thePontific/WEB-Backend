package ds

// — М-М заявки–услуги
type CartItem struct {
	ID       int     `gorm:"primaryKey"`
	CartID   int     `gorm:"not null;uniqueIndex:idx_cart_star"`
	StarID   int     `gorm:"not null;uniqueIndex:idx_cart_star"`
	Quantity int     `gorm:"default:1"`
	Speed    float32 // скорость
	Comment  string

	Cart *Cart `gorm:"foreignKey:CartID"`
	Star *Star `gorm:"foreignKey:StarID"`
}
