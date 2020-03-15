package app

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"crypto/tls"
	"database/sql"
	"gopkg.in/guregu/null.v3"
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

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods",
		"DELETE, GET, PATCH, POST, PUT")
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
