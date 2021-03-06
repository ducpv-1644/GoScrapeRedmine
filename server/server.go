package server

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go-scrape-redmine/server/handler"
	"gorm.io/gorm"
	"net/http"
	"strings"
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
		authorizationHeader := r.Header.Get("authorization")

		if authorizationHeader == "" {
			resp.Code = http.StatusBadRequest
			resp.Message = "No Token Found"
			handler.RespondWithJSON(w, resp.Code, resp)
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
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

func Run(db *gorm.DB) {
	router := mux.NewRouter()
	userHandler := handler.UserHandler{}
	userHandler.Db = db
	headersOk := handlers.AllowedHeaders([]string{"Accept", "content-type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Access-Control-Allow-Origin"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	router.Handle("/", isAuthorized(handlerx)).Methods("GET")
	router.HandleFunc("/signup", userHandler.SignUp).Methods("POST")
	router.HandleFunc("/signin", userHandler.SignIn).Methods("POST")
	//router.Handle("/activity", isAuthorized(userHandler.GetActivity)).Methods("GET")
	//router.Handle("/effort", isAuthorized(userHandler.GetEffort)).Methods("GET")
	router.Handle("/crawl", isAuthorized(userHandler.CrawData)).Methods("POST")
	router.Handle("/projects", isAuthorized(userHandler.GetAllProject)).Methods("GET")
	// router.Handle("/members", isAuthorized(user_handler.GetAllMember)).Methods("GET")
	//router.HandleFunc("/members", userHandler.GetAllMember).Methods("GET")
	//router.Handle("/member/{id}", isAuthorized(userHandler.GetAllIssue)).Methods("GET")
	router.Handle("/project_versions", isAuthorized(userHandler.GetAllVersionProject)).Methods("GET")
	router.Handle("/crawl_issues", isAuthorized(userHandler.CrawlIssueByVersion)).Methods("GET")
	router.Handle("/version_project", isAuthorized(userHandler.SetCurrentVersion)).Methods("POST")
	router.Handle("/config", isAuthorized(userHandler.CreateConfig)).Methods("POST")
	router.Handle("/config/{id}", isAuthorized(userHandler.UpdateConfig)).Methods("POST")
	router.Handle("/config", isAuthorized(userHandler.GetAllConfig)).Methods("GET")
	router.Handle("/config/{id}", isAuthorized(userHandler.GetConfigById)).Methods("GET")
	router.Handle("/config/{id}", isAuthorized(userHandler.DeleteConfig)).Methods("DELETE")
	router.Handle("/session", isAuthorized(userHandler.CreateSession)).Methods("POST")
	router.Handle("/session/{id}", isAuthorized(userHandler.GetSessionById)).Methods("GET")
	fmt.Println("Server started port 7000!")
	err := http.ListenAndServe(":7000", handlers.CORS(originsOk, headersOk, methodsOk)(router))
	fmt.Println("err", err)
	if err != nil {
		return
	}
}
