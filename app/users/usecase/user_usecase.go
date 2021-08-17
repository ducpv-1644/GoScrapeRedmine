package usecase

import (
	"go-scrape-redmine/app/users"
	"go-scrape-redmine/models"

	"gorm.io/gorm"
)

type userUsecase struct {
	userRepository        users.Repository
}

func NewUserUsecase(
	usersRepository        users.Repository,
) users.Usecase {
	return &userUsecase{
		usersRepository,
	}
}

func (usecase *userUsecase) FindUserByEmail(db *gorm.DB, email string) (models.User) {
	user := usecase.userRepository.FindUserByEmail(db, email)
	return user
}

func (usecase *userUsecase) CreateUser(db *gorm.DB, user models.User) {
	usecase.userRepository.CreateUser(db, user)
}
