package handler

import (
	"encoding/json"
	"go-scrape-redmine/app/users"
	userRepository "go-scrape-redmine/app/users/repository"
	userUsecase "go-scrape-redmine/app/users/usecase"
	"go-scrape-redmine/config"
	"go-scrape-redmine/models"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct{}
type response struct {
        Code    int  `json:"code"`
        Message string `json:"message"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func newUserUsecase() users.Usecase {
	return userUsecase.NewUserUsecase(
		userRepository.NewUserRepository(),
	)
}

func generatehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (a *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	db := config.DBConnect()

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		respondWithJSON(w, resp.Code, resp)
		return
	}

	dbuser := newUserUsecase().FindUserByEmail(db, user.Email)

	//checks if email is already register or not
	if dbuser.Email != "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Email already in use!"
		respondWithJSON(w, resp.Code, resp)
		return
	}

	user.Password, err = generatehashPassword(user.Password)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = "error in password hash!"
		respondWithJSON(w, resp.Code, resp)
		return
	}

	newUserUsecase().CreateUser(db, user)

	respondWithJSON(w, http.StatusOK, user)
}
