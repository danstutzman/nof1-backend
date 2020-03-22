package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/model"
	"crypto/tls"
	"database/sql"
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"os"
	"time"
)

type WebApp struct {
	model     *model.Model
	dbConn    *sql.DB
	staticDir string
}

func NewWebApp(
	model *model.Model,
	dbConn *sql.DB,
	staticDir string,
) *WebApp {
	return &WebApp{
		model:     model,
		dbConn:    dbConn,
		staticDir: staticDir,
	}
}

func (webapp *WebApp) logRequest(receivedAt time.Time, r *http.Request,
	statusCode, size int, errorStack null.String,
	browser *db.BrowsersRow) db.RequestsRow {

	var tlsProtocol null.String
	var tlsCipher null.String
	if r.TLS != nil {
		tlsProtocol = null.StringFrom(r.TLS.NegotiatedProtocol)
		tlsCipher = null.StringFrom(tls.CipherSuiteName(r.TLS.CipherSuite))
	}

	log.Printf("%s %s\n", r.Method, r.URL.RequestURI())

	var browserId null.Int
	if browser != nil {
		browserId = null.IntFrom(browser.Id)
		db.TouchBrowserLastSeenAt(webapp.dbConn, browser.Id)
	}

	return db.InsertIntoRequests(webapp.dbConn, db.RequestsRow{
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

func (webapp *WebApp) getRoot(r *http.Request,
	browser *db.BrowsersRow) Response {

	path := webapp.staticDir + "/index.html"
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ErrorResponse{http.StatusNotFound}
	} else if err != nil {
		panic(err)
	}

	return FileResponse{path: path, size: int(fileInfo.Size()), mimeType: ""}
}

func (webapp *WebApp) getStaticFile(r *http.Request,
	browser *db.BrowsersRow) Response {

	path := webapp.staticDir + r.URL.RequestURI()

	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ErrorResponse{status: http.StatusNotFound}
	} else if err != nil {
		panic(err)
	}

	return FileResponse{path: path, size: int(fileInfo.Size()), mimeType: ""}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods",
		"DELETE, GET, PATCH, POST, PUT")
}
