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
	"strconv"
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

func getBrowserIdCookie(w http.ResponseWriter, r *http.Request,
	secretKey string) int {

	cookie, err := r.Cookie("browser-id")
	if err == nil {
		decrypted, err := decrypt(cookie.Value, secretKey)
		if err != nil {
			log.Printf("Couldn't decrypt cookie: %v", err)
			http.SetCookie(w, &http.Cookie{
				Name:    "browser-id",
				Expires: time.Unix(0, 0),
			})
			return 0
		}
		browserId, _ := strconv.Atoi(decrypted)
		return browserId
	} else if err == http.ErrNoCookie {
		return 0
	} else {
		panic(err)
	}
}

func getOrSetBrowserIdCookie(w http.ResponseWriter, r *http.Request,
	dbConn *sql.DB, secretKey string) int {
	cookie, err := r.Cookie("browser-id")
	if err == nil {
		decrypted, err := decrypt(cookie.Value, secretKey)
		if err != nil {
			log.Printf("Couldn't decrypt cookie: %v", err)
			http.SetCookie(w, &http.Cookie{
				Name:    "browser-id",
				Expires: time.Unix(0, 0),
			})
			return 0
		}
		browserId, _ := strconv.Atoi(decrypted)
		return browserId
	} else if err == http.ErrNoCookie {
		browser := db.InsertIntoBrowsers(dbConn, db.BrowsersRow{
			UserAgent:      r.UserAgent(),
			Accept:         r.Header.Get("Accept"),
			AcceptEncoding: r.Header.Get("Accept-Encoding"),
			AcceptLanguage: r.Header.Get("Accept-Language"),
			Referer:        r.Referer(),
		})

		http.SetCookie(w, &http.Cookie{
			Name:    "browser-id",
			Value:   encrypt(strconv.Itoa(browser.Id), secretKey),
			Expires: time.Now().AddDate(30, 0, 0),
		})

		return browser.Id
	} else {
		panic(err)
	}
}

func logRequest(dbConn *sql.DB, receivedAt time.Time, r *http.Request,
	statusCode, size int, errorStack null.String, browserId int) db.RequestsRow {

	var tlsProtocol null.String
	var tlsCipher null.String
	if r.TLS != nil {
		tlsProtocol = null.StringFrom(r.TLS.NegotiatedProtocol)
		tlsCipher = null.StringFrom(tls.CipherSuiteName(r.TLS.CipherSuite))
	}

	log.Printf("%s %s\n", r.Method, r.URL.RequestURI())

	return db.InsertIntoRequests(dbConn, db.RequestsRow{
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

func getRoot(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	staticDir string, secretKey string) {
	receivedAt := time.Now().UTC()
	browserId := getOrSetBrowserIdCookie(w, r, dbConn, secretKey)

	html, err := ioutil.ReadFile(staticDir + "/index.html")
	if os.IsNotExist(err) {
		notFound(w, r, dbConn, secretKey)
		return
	} else if err != nil {
		panic(err)
	}
	w.Write([]byte(html))

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(html), null.String{},
		browserId)
}

func getStaticFile(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	staticDir string, secretKey string) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, secretKey)

	bytes, err := ioutil.ReadFile(staticDir + r.URL.RequestURI())
	if os.IsNotExist(err) {
		notFound(w, r, dbConn, secretKey)
		return
	} else if err != nil {
		panic(err)
	}
	w.Write([]byte(bytes))

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}

func postUploadAudio(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	secretKey string) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, secretKey)

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

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(bytes), null.String{},
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

func postSync(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	secretKey string) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, secretKey)
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
		db.InsertIntoLogs(dbConn, convertClientLogToLogsRow(clientLog, browserId))
	}

	bytes := "{}"
	w.Write([]byte(bytes))

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(bytes), null.String{},
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

func notFound(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	secretKey string) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, secretKey)

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

	logRequest(dbConn, receivedAt, r, http.StatusNotFound,
		len(http.StatusText(http.StatusNotFound)), null.String{}, browserId)
}

func getWithoutTls(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	secretKey string) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r, secretKey)

	http.Redirect(w, r, "https://wellsaid.us"+r.RequestURI,
		http.StatusMovedPermanently)

	logRequest(
		dbConn, receivedAt, r, http.StatusMovedPermanently, 0, null.String{},
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

	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			notFound(w, r, dbConn, secretKey)
		})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		getRoot(w, r, dbConn, staticDir, secretKey)
	})
	router.HandleFunc("/{prefix}.mp3",
		func(w http.ResponseWriter, r *http.Request) {
			getStaticFile(w, r, dbConn, staticDir, secretKey)
		})
	router.HandleFunc("/bundle.js",
		func(w http.ResponseWriter, r *http.Request) {
			getStaticFile(w, r, dbConn, staticDir, secretKey)
		})
	router.HandleFunc("/bundle.js.map",
		func(w http.ResponseWriter, r *http.Request) {
			getStaticFile(w, r, dbConn, staticDir, secretKey)
		})
	router.HandleFunc("/upload-audio",
		func(w http.ResponseWriter, r *http.Request) {
			postUploadAudio(w, r, dbConn, secretKey)
		})
	router.HandleFunc("/sync",
		func(w http.ResponseWriter, r *http.Request) {
			postSync(w, r, dbConn, secretKey)
		})

	if httpsCertFile != "" || httpsKeyFile != "" {
		log.Printf("Serving TLS on :443 and HTTP on :" + httpPort + "...")

		go func() {
			err := http.ListenAndServeTLS(":443", httpsCertFile, httpsKeyFile,
				newRecoveryHandler(router, dbConn, secretKey))
			panic(err)
		}()

		redirectToTlsRouter := mux.NewRouter()
		redirectToTlsRouter.NotFoundHandler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				getWithoutTls(w, r, dbConn, secretKey)
			})
		err := http.ListenAndServe(":"+httpPort,
			newRecoveryHandler(redirectToTlsRouter, dbConn, secretKey))
		panic(err)
	} else {
		log.Printf("Serving HTTP on :" + httpPort + "...")
		err := http.ListenAndServe(":"+httpPort,
			newRecoveryHandler(router, dbConn, secretKey))
		panic(err)
	}
}
