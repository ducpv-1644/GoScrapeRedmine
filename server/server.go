package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
        "go-scrape-redmine/server/handler"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type response struct {
        Code    int  `json:"code"`
        Message string `json:"message"`
}

func handlerx(w http.ResponseWriter, r *http.Request) {
        resp := response{}
        resp.Code = http.StatusOK
        resp.Message = "Hello World!"
        RespondWithJSON(w, resp.Code, resp)
}

var mySigningKey = []byte("unicorns")

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	resp := response{}

        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                if r.Header["Token"] != nil {
                        token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
                                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                                        return nil, fmt.Errorf("Token invalid!")
                                }
                                return mySigningKey, nil
                        })

                        if err != nil {
                                resp.Code = http.StatusBadRequest
                                resp.Message = err.Error()
                                RespondWithJSON(w, resp.Code, resp)
                                return
                        }
                        if token.Valid {
                                endpoint(w, r)
                                return
                        }
                } else {
                        resp.Code = http.StatusNonAuthoritativeInfo
                        resp.Message = "Not Authorized!"
                        RespondWithJSON(w, resp.Code, resp)
                }
        })
}

func Run(wg *sync.WaitGroup) {
        router := mux.NewRouter()
        user_handler := handler.UserHandler{}
        defer wg.Done()

        router.Handle("/", isAuthorized(handlerx)).Methods("GET")
        router.HandleFunc("/signup", user_handler.SignUp).Methods("POST")

        fmt.Println("Server started port 8000!")
        http.ListenAndServe(":8000", router)
}
