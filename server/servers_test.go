package server

import (
	"bytes"

	"github.com/stretchr/testify/mock"

	"go-scrape-redmine/server/handler"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRepository struct {
	mock.Mock
}

func (d *mockRepository) FetchMessage(email string, password string) (string, string) {
	args := d.Called(email, password)

	return args.String(0), args.String(1)
}
func TestSignIn(t *testing.T) {

	theDBMock := mockRepository{}
	theDBMock.On("FetchMessage").Return("", "")

	user_handler := handler.UserHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/signin", user_handler.SignIn)
	var jsonStr = []byte(` {
		 "email": "",
		 "password": ""
	     }`)

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/signin", bytes.NewBuffer(jsonStr))
	request.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(writer, request)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}
}
