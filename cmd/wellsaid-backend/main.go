package main

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/guregu/null.v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type SyncRequest struct {
	Logs []map[string]interface{}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods",
		"DELETE, GET, PATCH, POST, PUT")
}

type App struct {
	dbConn    *sql.DB
	staticDir string
	secretKey string
}

func (app *App) logRequest(receivedAt time.Time, r *http.Request,
	statusCode, size int, errorStack null.String, browserId int) db.RequestsRow {

	var tlsProtocol null.String
	var tlsCipher null.String
	if r.TLS != nil {
		tlsProtocol = null.StringFrom(r.TLS.NegotiatedProtocol)
		tlsCipher = null.StringFrom(tls.CipherSuiteName(r.TLS.CipherSuite))
	}

	log.Printf("%s %s\n", r.Method, r.URL.RequestURI())

	return db.InsertIntoRequests(app.dbConn, db.RequestsRow{
		ReceivedAt:  receivedAt,
		RemoteAddr:  r.RemoteAddr,
		BrowserId:   browserId,
		HttpVersion: r.Proto,
		TlsProtocol: tlsProtocol,
		TlsCipher:   tlsCipher,
		Method:      r.Method,
		Path:        r.URL.RequestURI(),
		DurationMs:  int(time.Now().UTC().Sub(receivedAt).Milliseconds()),
		StatusCode:  statusCode,
		Size:        size,
		ErrorStack:  errorStack,
	})
}

func (app *App) getRoot(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := getOrSetBrowserIdCookie(w, r, app.dbConn, app.secretKey)

	html, err := ioutil.ReadFile(app.staticDir + "/index.html")
	if os.IsNotExist(err) {
		app.notFound(w, r)
		return
	} else if err != nil {
		panic(err)
	}
	w.Write([]byte(html))

	app.logRequest(receivedAt, r, http.StatusOK, len(html), null.String{},
		browserId)
}

func (app *App) getStaticFile(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, app.secretKey)

	bytes, err := ioutil.ReadFile(app.staticDir + r.URL.RequestURI())
	if os.IsNotExist(err) {
		app.notFound(w, r)
		return
	} else if err != nil {
		panic(err)
	}
	w.Write([]byte(bytes))

	app.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}

func (app *App) postUploadAudio(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, app.secretKey)

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("audio_data")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.Printf("Header: %v", handler.Header)

	f, err := os.OpenFile("/tmp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.Copy(f, file)

	bytes := "OK"
	w.Write([]byte(bytes))

	app.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}

func convertClientLogToLogsRow(clientLog map[string]interface{},
	browserId int) db.LogsRow {

	var idOnClient int
	if f, ok := clientLog["id"].(float64); ok {
		idOnClient = int(f)
	}
	delete(clientLog, "id")

	var timeOnClient int
	if f, ok := clientLog["time"].(float64); ok {
		timeOnClient = int(f)
	}
	delete(clientLog, "time")

	message := clientLog["message"].(string)
	delete(clientLog, "message")

	var errorName null.String
	var errorMessage null.String
	var errorStack null.String
	if clientLog["error"] != nil {
		if errorMap, ok := clientLog["error"].(map[string]interface{}); ok {
			if s, ok := errorMap["name"].(string); ok {
				errorName = null.StringFrom(s)
			}
			if s, ok := errorMap["message"].(string); ok {
				errorMessage = null.StringFrom(s)
			}
			if s, ok := errorMap["stack"].(string); ok {
				errorStack = null.StringFrom(s)
			}
			delete(clientLog, "error")
		}
	}

	var otherDetailsJson null.String
	if len(clientLog) > 0 {
		json, err := json.Marshal(clientLog)
		if err != nil {
			panic(err)
		}
		otherDetailsJson = null.StringFrom(string(json))
	}

	return db.LogsRow{
		BrowserId:        browserId,
		IdOnClient:       idOnClient,
		TimeOnClient:     timeOnClient,
		Message:          message,
		ErrorName:        errorName,
		ErrorMessage:     errorMessage,
		ErrorStack:       errorStack,
		OtherDetailsJson: otherDetailsJson,
	}
}

func (app *App) postSync(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, app.secretKey)
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=\"utf-8\"")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var syncRequest SyncRequest
	err = json.Unmarshal(body, &syncRequest)
	if err != nil {
		panic(err)
	}

	for _, clientLog := range syncRequest.Logs {
		db.InsertIntoLogs(app.dbConn,
			convertClientLogToLogsRow(clientLog, browserId))
	}

	bytes := "{}"
	w.Write([]byte(bytes))

	app.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}

func param(r *http.Request, key string) null.String {
	values := r.Form[key]
	if len(values) == 1 {
		return null.StringFrom(values[0])
	} else {
		return null.String{}
	}
}

func (app *App) notFound(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, app.secretKey)

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

	app.logRequest(receivedAt, r, http.StatusNotFound,
		len(http.StatusText(http.StatusNotFound)), null.String{}, browserId)
}

func (app *App) getWithoutTls(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, app.secretKey)

	http.Redirect(w, r, "https://wellsaid.us"+r.RequestURI,
		http.StatusMovedPermanently)

	app.logRequest(receivedAt, r, http.StatusMovedPermanently, 0, null.String{},
		browserId)
}

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
