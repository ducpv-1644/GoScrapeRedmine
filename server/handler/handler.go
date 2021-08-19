package handler

import (
	"encoding/json"
	"fmt"
	"go-scrape-redmine/app/users"
	userRepository "go-scrape-redmine/app/users/repository"
	userUsecase "go-scrape-redmine/app/users/usecase"
	"go-scrape-redmine/config"
	"go-scrape-redmine/crawl"
	"go-scrape-redmine/models"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
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

func generateJWT(email, role string) (string, error) {
	var mySigningKey = []byte("unicorns")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = email
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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

	userCreated := newUserUsecase().CreateUser(db, user)
	respondWithJSON(w, http.StatusOK, userCreated)
}

func (a *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	db := config.DBConnect()

	var authdetails models.Authentication
	err := json.NewDecoder(r.Body).Decode(&authdetails)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		respondWithJSON(w, resp.Code, resp)
		return
	}

	authuser := newUserUsecase().FindUserByEmail(db, authdetails.Email)
	passwordOK := checkPasswordHash(authdetails.Password, authuser.Password)
	if !passwordOK || authuser.Email == "" {
		resp.Code = http.StatusBadRequest
		resp.Message = "Username or Password is incorrect"
		respondWithJSON(w, resp.Code, resp)
		return
	}

	validToken, err := generateJWT(authuser.Email, authuser.Role)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = "Failed to generate token"
		respondWithJSON(w, resp.Code, resp)
		return
	}

	var token models.Token
	token.Email = authuser.Email
	token.Role = authuser.Role
	token.TokenString = validToken
	respondWithJSON(w, http.StatusOK, token)
}

func (a *UserHandler) CrawRedmineData(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	crawl.CrawlData()

	resp.Code = http.StatusOK
	resp.Message = "Request crawl redmine data finished."
	respondWithJSON(w, resp.Code, resp)
}
