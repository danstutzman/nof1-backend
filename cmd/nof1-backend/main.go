package main

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	modelPkg "bitbucket.org/danstutzman/nof1-backend/internal/model"
	webappPkg "bitbucket.org/danstutzman/nof1-backend/internal/webapp"
	"github.com/NYTimes/gziphandler"
	"log"
	"net/http"
	"os"
)

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		log.Fatalf("Set HTTP_PORT env var")
	}

	httpsCertFile := os.Getenv("HTTPS_CERT_FILE")
	httpsKeyFile := os.Getenv("HTTPS_KEY_FILE")

	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		log.Fatalf("Set DB_FILE env var")
	}
	dbConn := db.InitDb(dbFile)

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		log.Fatalf("Set STATIC_DIR env var")
	}

	model := modelPkg.NewModel(dbConn, "/tmp/nof1-backend")
	webapp := webappPkg.NewWebApp(model, dbConn, staticDir)
	router := gziphandler.GzipHandler(webappPkg.NewRouter(webapp))
	redirectToTlsRouter := webappPkg.NewRedirectToTlsRouter(webapp)

	if httpsCertFile != "" || httpsKeyFile != "" {
		log.Printf("Serving TLS on :443 and HTTP on :" + httpPort + "...")

		go func() {
			err := http.ListenAndServeTLS(":443", httpsCertFile, httpsKeyFile,
				router)
			panic(err)
		}()

		err := http.ListenAndServe(":"+httpPort, redirectToTlsRouter)
		panic(err)
	} else {
		log.Printf("Serving HTTP on :" + httpPort + "...")
		err := http.ListenAndServe(":"+httpPort, router)
		panic(err)
	}
}
