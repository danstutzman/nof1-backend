package main

import (
	appPkg "bitbucket.org/danstutzman/wellsaid-backend/internal/app"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
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

	app := appPkg.NewApp(dbConn, staticDir)
	router := appPkg.NewRouter(app)
	redirectToTlsRouter := appPkg.NewRedirectToTlsRouter(app)

	if httpsCertFile != "" || httpsKeyFile != "" {
		log.Printf("Serving TLS on :443 and HTTP on :" + httpPort + "...")

		go func() {
			err := http.ListenAndServeTLS(":443", httpsCertFile, httpsKeyFile,
				appPkg.NewRecoveryHandler(router, app))
			panic(err)
		}()

		err := http.ListenAndServe(":"+httpPort,
			appPkg.NewRecoveryHandler(redirectToTlsRouter, app))
		panic(err)
	} else {
		log.Printf("Serving HTTP on :" + httpPort + "...")
		err := http.ListenAndServe(":"+httpPort,
			appPkg.NewRecoveryHandler(router, app))
		panic(err)
	}
}
