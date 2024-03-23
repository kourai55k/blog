package data

import (
	"gorm.io/gorm"
)

// Команды для таблицы users

func CreateUser(db *gorm.DB, user User) error {
	res := db.Create(&user)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// GetUser возвращает пользователя из БД по ID
//func GetUser(db *gorm.DB, id int) (User, error) {
//	var user User
//	res := db.First(&user, id)
//	return user, res.Error
//}

// GetUserByLogin возвращает пользователя из БД по ID
func GetUserByLogin(db *gorm.DB, login string) (User, error) {
	var user User
	res := db.Where("login = ?", login).First(&user)
	return user, res.Error
}

// Команды для таблицы posts

// CreatePost добавляет пост в базу данных
func CreatePost(db *gorm.DB, post Post) error {
	res := db.Create(&post)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// GetPost возвращает пост из БД по ID
func GetPost(db *gorm.DB, id int) (Post, error) {
	var post Post
	res := db.First(&post, id)
	return post, res.Error
}

// GetPosts возвращает все посты из БД
func GetPosts(db *gorm.DB) ([]Post, error) {
	posts := make([]Post, 0)
	res := db.Find(&posts)
	return posts, res.Error
}

func DeletePost(db *gorm.DB, id int) error {
	res := db.Delete(&Post{}, id)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
