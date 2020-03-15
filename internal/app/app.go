package app

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"gopkg.in/guregu/null.v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type App struct {
	dbConn    *sql.DB
	staticDir string
	secretKey string
}

func NewApp(
	dbConn *sql.DB,
	staticDir string,
	secretKey string,
) *App {
	return &App{
		dbConn:    dbConn,
		staticDir: staticDir,
		secretKey: secretKey,
	}
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

func (app *App) GetRoot(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getOrSetBrowserIdCookie(w, r)

	html, err := ioutil.ReadFile(app.staticDir + "/index.html")
	if os.IsNotExist(err) {
		app.NotFound(w, r)
		return
	} else if err != nil {
		panic(err)
	}
	w.Write([]byte(html))

	app.logRequest(receivedAt, r, http.StatusOK, len(html), null.String{},
		browserId)
}

func (app *App) GetStaticFile(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getBrowserIdCookie(w, r)

	bytes, err := ioutil.ReadFile(app.staticDir + r.URL.RequestURI())
	if os.IsNotExist(err) {
		app.NotFound(w, r)
		return
	} else if err != nil {
		panic(err)
	}
	w.Write([]byte(bytes))

	app.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}

func (app *App) PostUploadAudio(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getBrowserIdCookie(w, r)

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

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods",
		"DELETE, GET, PATCH, POST, PUT")
}

type SyncRequest struct {
	Logs []map[string]interface{}
}

func (app *App) PostSync(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getBrowserIdCookie(w, r)
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

func (app *App) NotFound(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getBrowserIdCookie(w, r)

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

	app.logRequest(receivedAt, r, http.StatusNotFound,
		len(http.StatusText(http.StatusNotFound)), null.String{}, browserId)
}

func (app *App) GetWithoutTls(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getBrowserIdCookie(w, r)

	http.Redirect(w, r, "https://wellsaid.us"+r.RequestURI,
		http.StatusMovedPermanently)

	app.logRequest(receivedAt, r, http.StatusMovedPermanently, 0, null.String{},
		browserId)
}
