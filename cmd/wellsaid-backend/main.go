package main

import (
	appPkg "bitbucket.org/danstutzman/wellsaid-backend/internal/app"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	webappPkg "bitbucket.org/danstutzman/wellsaid-backend/internal/webapp"
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

	app := appPkg.NewApp(dbConn, "/tmp")
	webapp := webappPkg.NewWebApp(app, dbConn, staticDir)
	router := webappPkg.NewRouter(webapp)
	redirectToTlsRouter := webappPkg.NewRedirectToTlsRouter(webapp)

	if httpsCertFile != "" || httpsKeyFile != "" {
		log.Printf("Serving TLS on :443 and HTTP on :" + httpPort + "...")

		go func() {
			err := http.ListenAndServeTLS(":443", httpsCertFile, httpsKeyFile,
				webappPkg.NewRecoveryHandler(router, webapp))
			panic(err)
		}()

		err := http.ListenAndServe(":"+httpPort,
			webappPkg.NewRecoveryHandler(redirectToTlsRouter, webapp))
		panic(err)
	} else {
		log.Printf("Serving HTTP on :" + httpPort + "...")
		err := http.ListenAndServe(":"+httpPort,
			webappPkg.NewRecoveryHandler(router, webapp))
		panic(err)
	}
}
