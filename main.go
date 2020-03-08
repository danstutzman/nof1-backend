package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		log.Fatalf("Set HTTP_PORT env var")
	}

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a catch-all route\n"))
	})
	loggedRouter := handlers.CombinedLoggingHandler(os.Stderr, router)

	log.Printf("Listening on :" + httpPort + "...")
	err := http.ListenAndServe(":"+httpPort, loggedRouter)
	panic(err)
}
