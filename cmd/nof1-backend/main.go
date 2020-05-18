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

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatalf("Set ADMIN_PASSWORD env var")
	}

	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyId == "" {
		log.Fatalf("Set AWS_ACCESS_KEY_ID env var")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatalf("Set AWS_REGION env var")
	}

	awsS3Bucket := os.Getenv("AWS_S3_BUCKET")
	if awsS3Bucket == "" {
		log.Fatalf("Set AWS_S3_BUCKET env var")
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		log.Fatalf("Set AWS_SECRET_ACCESS_KEY env var")
	}

	model := modelPkg.NewModel(modelPkg.Config{
		AwsAccessKeyId:     awsAccessKeyId,
		AwsRegion:          awsRegion,
		AwsS3Bucket:        awsS3Bucket,
		AwsSecretAccessKey: awsSecretAccessKey,
		DbConn:             dbConn,
		UploadDir:          "/tmp/nof1-backend",
	})
	webapp := webappPkg.NewWebApp(model, dbConn, staticDir, adminPassword)
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
