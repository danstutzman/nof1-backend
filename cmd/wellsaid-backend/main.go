package main

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	secretKey := os.Getenv("SECRET_KEY")
	if !isSecretKeyOkay(secretKey) {
		log.Fatalf("Set SECRET_KEY env var to any random 32 bytes Base64-encoded, "+
			"for example: %s", makeExampleSecretKey())
	}

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

	app := App{
		dbConn:    dbConn,
		staticDir: staticDir,
		secretKey: secretKey,
	}

	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			app.notFound(w, r)
		})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		app.getRoot(w, r)
	})
	router.HandleFunc("/{prefix}.mp3",
		func(w http.ResponseWriter, r *http.Request) {
			app.getStaticFile(w, r)
		})
	router.HandleFunc("/bundle.js",
		func(w http.ResponseWriter, r *http.Request) {
			app.getStaticFile(w, r)
		})
	router.HandleFunc("/bundle.js.map",
		func(w http.ResponseWriter, r *http.Request) {
			app.getStaticFile(w, r)
		})
	router.HandleFunc("/upload-audio",
		func(w http.ResponseWriter, r *http.Request) {
			app.postUploadAudio(w, r)
		})
	router.HandleFunc("/sync",
		func(w http.ResponseWriter, r *http.Request) {
			app.postSync(w, r)
		})

	if httpsCertFile != "" || httpsKeyFile != "" {
		log.Printf("Serving TLS on :443 and HTTP on :" + httpPort + "...")

		go func() {
			err := http.ListenAndServeTLS(":443", httpsCertFile, httpsKeyFile,
				newRecoveryHandler(router, app))
			panic(err)
		}()

		redirectToTlsRouter := mux.NewRouter()
		redirectToTlsRouter.NotFoundHandler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				app.getWithoutTls(w, r)
			})
		err := http.ListenAndServe(":"+httpPort,
			newRecoveryHandler(redirectToTlsRouter, app))
		panic(err)
	} else {
		log.Printf("Serving HTTP on :" + httpPort + "...")
		err := http.ListenAndServe(":"+httpPort, newRecoveryHandler(router, app))
		panic(err)
	}
}
