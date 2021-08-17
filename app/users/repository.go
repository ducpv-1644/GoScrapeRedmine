package users

import (
	"go-scrape-redmine/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindUserByEmail(db *gorm.DB, email string) (models.User)
	CreateUser(db *gorm.DB, user models.User)
}
