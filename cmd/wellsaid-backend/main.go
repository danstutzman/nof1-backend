package main

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"crypto/tls"
	"database/sql"
	"github.com/gorilla/mux"
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"os"
	"time"
)

func logRequest(dbConn *sql.DB, receivedAt time.Time, r *http.Request, statusCode, size int) {

	var tlsProtocol null.String
	var tlsCipher null.String
	if r.TLS != nil {
		tlsProtocol = null.StringFrom(r.TLS.NegotiatedProtocol)
		tlsCipher = null.StringFrom(tls.CipherSuiteName(r.TLS.CipherSuite))
	}

	db.InsertIntoRequests(dbConn, db.RequestsRow{
		ReceivedAt:  receivedAt,
		RemoteAddr:  r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Referer:     r.Referer(),
		HttpVersion: r.Proto,
		TlsProtocol: tlsProtocol,
		TlsCipher:   tlsCipher,
		Method:      r.Method,
		Path:        r.URL.RequestURI(),
		DurationMs:  int(time.Now().UTC().Sub(receivedAt).Milliseconds()),
		StatusCode:  statusCode,
		Size:        size,
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

	log.Printf("protocol=%s", r.TLS.NegotiatedProtocol)

	output := "This is a catch-all route\n"
	w.Write([]byte(output))

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(output))
}

func handleNotFound(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

	output := "Not found\n"
	w.Write([]byte(output))

	logRequest(dbConn, receivedAt, r, http.StatusNotFound, len(output))
}

func handleRedirectToTls(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

	http.Redirect(w, r, "https://wellsaid.us"+r.RequestURI,
		http.StatusMovedPermanently)

	logRequest(dbConn, receivedAt, r, http.StatusMovedPermanently, 0)
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		log.Fatalf("Set HTTP_PORT env var")
	}

	httpsCertFile := os.Getenv("HTTPS_CERT_FILE")
	httpsKeyFile := os.Getenv("HTTPS_KEY_FILE")

	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		log.Fatalf("Set DB_FILE env var.")
	}
	dbConn := db.InitDb(dbFile)

	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handleNotFound(w, r, dbConn)
		})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleIndex(w, r, dbConn)
	})

	if httpsCertFile != "" || httpsKeyFile != "" {
		log.Printf("Serving TLS on :443 and HTTP on :" + httpPort + "...")

		go func() {
			err := http.ListenAndServeTLS(":443", httpsCertFile, httpsKeyFile, router)
			panic(err)
		}()

		redirectToTlsRouter := mux.NewRouter()
		redirectToTlsRouter.NotFoundHandler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				handleRedirectToTls(w, r, dbConn)
			})
		err := http.ListenAndServe(":"+httpPort, redirectToTlsRouter)
		panic(err)
	} else {
		log.Printf("Serving HTTP on :" + httpPort + "...")
		err := http.ListenAndServe(":"+httpPort, router)
		panic(err)
	}
}
