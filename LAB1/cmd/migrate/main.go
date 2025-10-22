package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"LAB1/internal/app/ds"
	"LAB1/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema: только 4 таблицы
	err = db.AutoMigrate(
		&ds.Star{},         // услуги
		&ds.StarCart{},     // заявки
		&ds.StarCartItem{}, // М-М заявки–услуги
		&ds.User{},         // пользователи
	)
	if err != nil {
		panic("cant migrate db")
	}
}
