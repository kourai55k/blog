package data

import (
	"gorm.io/gorm"
	"time"
)

type Post struct {
	Id        uint   `gorm:"primaryKey"`
	Author    string `gorm:"not null"`
	UserId    uint   `gorm:"not null"`
	Title     string `gorm:"not null"`
	Body      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	Id         uint   `gorm:"primaryKey"`
	Name       string `gorm:"not null"`
	Login      string `gorm:"not null;unique"`
	HashedPass string `gorm:"not null"`
}

// CreateTables Создаёт таблицы в БД
func CreateTables(db *gorm.DB) error {
	err := db.AutoMigrate(&Post{}, &User{})
	if err != nil {
		return err
	}
	return nil
}
