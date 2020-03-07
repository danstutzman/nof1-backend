package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a catch-all route"))
	})
	loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)

	log.Printf("Listening on :8080...")
	err := http.ListenAndServe(":8080", loggedRouter)
	panic(err)
}
