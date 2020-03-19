package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/model"
	"crypto/tls"
	"database/sql"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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

	html, err := ioutil.ReadFile(webapp.staticDir + "/index.html")
	if os.IsNotExist(err) {
		return ErrorResponse{http.StatusNotFound}
	} else if err != nil {
		panic(err)
	}

	return BytesResponse{content: html}
}

func (webapp *WebApp) getStaticFile(r *http.Request,
	browser *db.BrowsersRow) Response {
	filename := r.URL.RequestURI()

	bytes, err := ioutil.ReadFile(webapp.staticDir + filename)
	if os.IsNotExist(err) {
		return ErrorResponse{status: http.StatusNotFound}
	} else if err != nil {
		panic(err)
	}

	var contentType string
	if strings.HasSuffix(filename, ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(filename, ".js") {
		contentType = "text/javascript"
	} else if strings.HasSuffix(filename, ".js.map") {
		contentType = "application/json" // so can view source in browser
	} else {
		panic("Unknown mime type")
	}

	return BytesResponse{content: bytes, contentType: contentType}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods",
		"DELETE, GET, PATCH, POST, PUT")
}

func (webapp *WebApp) notFound(r *http.Request,
	browser *db.BrowsersRow) Response {

	return ErrorResponse{status: http.StatusNotFound}
}

func (webapp *WebApp) getWithoutTls(r *http.Request,
	browser *db.BrowsersRow) Response {

	return RedirectResponse{url: "https://wellsaid.us" + r.RequestURI}
}
