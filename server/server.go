package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

func handler(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello World!")
}

func Run(wg *sync.WaitGroup) {
        router := mux.NewRouter()

        defer wg.Done()

        router.HandleFunc("/", handler).Methods("GET")

        fmt.Println("Server started port 8000!")
        http.ListenAndServe(":8000", router)
}
