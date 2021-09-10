package server

import (
	"fmt"
	"go-scrape-redmine/server/handler"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handlerx(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	resp.Code = http.StatusOK
	resp.Message = "Hello World!"
	handler.RespondWithJSON(w, resp.Code, resp)
}

var mySigningKey = []byte("unicorns")

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	resp := response{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] == nil {
			resp.Code = http.StatusBadRequest
			resp.Message = "No Token Found"
			handler.RespondWithJSON(w, resp.Code, resp)
			return
		}
		token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Token invalid!")
			}
			return mySigningKey, nil
		})
		if err != nil {
			resp.Code = http.StatusBadRequest
			resp.Message = "Your Token has been expired"
			handler.RespondWithJSON(w, resp.Code, resp)
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["role"] == "admin" {
				r.Header.Set("Role", "admin")
				endpoint(w, r)
				return

			} else if claims["role"] == "user" {
				r.Header.Set("Role", "user")
				endpoint(w, r)
				return
			}
		}
		resp.Code = http.StatusNonAuthoritativeInfo
		resp.Message = "Not Authorized!"
		handler.RespondWithJSON(w, resp.Code, resp)
	})
}

func Run(wg *sync.WaitGroup) {
	router := mux.NewRouter()
	user_handler := handler.UserHandler{}
	defer wg.Done()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	router.Handle("/", isAuthorized(handlerx)).Methods("GET")
	router.HandleFunc("/signup", user_handler.SignUp).Methods("POST")
	router.HandleFunc("/signin", user_handler.SignIn).Methods("POST")
	router.Handle("/activity", isAuthorized(user_handler.GetActivity)).Methods("GET")
	router.Handle("/effort", isAuthorized(user_handler.GetEffort)).Methods("GET")
	router.Handle("/crawl", isAuthorized(user_handler.CrawData)).Methods("POST")
	router.Handle("/projects", isAuthorized(user_handler.GetAllProject)).Methods("GET")

	router.HandleFunc("/projects", user_handler.GetAllProject).Methods("GET")

	fmt.Println("Server started port 8000!")
	http.ListenAndServe(":8000", handlers.CORS(originsOk, headersOk, methodsOk)(router))
}
