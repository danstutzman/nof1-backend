package main

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		log.Fatalf("Set HTTP_PORT env var")
	}

	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		log.Fatalf("Set DB_FILE env var.")
	}
	dbConn := db.InitDb(dbFile)

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		receivedAt := time.Now().UTC()

		output := "This is a catch-all route\n"
		w.Write([]byte(output))

		db.InsertIntoRequests(dbConn, db.RequestsRow{
			ReceivedAt: receivedAt,
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			Referer:    r.Referer(),
			Method:     r.Method,
			Path:       r.URL.RequestURI(),
			StatusCode: 200,
			Size:       len(output),
		})
	})

	log.Printf("Listening on :" + httpPort + "...")
	err := http.ListenAndServe(":"+httpPort, router)
	panic(err)
}
