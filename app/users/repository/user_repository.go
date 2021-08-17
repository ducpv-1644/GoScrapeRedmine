package repository

import (
	"go-scrape-redmine/app/users"
	"go-scrape-redmine/models"

	"gorm.io/gorm"
)

func NewUserRepository() users.Repository {
	return &UserRepository{}
}

type UserRepository struct{}

func (*UserRepository) FindUserByEmail(db *gorm.DB, email string) (models.User){
	var dbuser models.User
	db.Where("email = ?", email).First(&dbuser)
	return dbuser
}

func (*UserRepository) CreateUser(db *gorm.DB, user models.User){
	db.Create(&user)
}
