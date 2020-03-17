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

func (webapp *WebApp) getRoot(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)

	html, err := ioutil.ReadFile(webapp.staticDir + "/index.html")
	if os.IsNotExist(err) {
		webapp.notFound(w, r)
		return
	} else if err != nil {
		panic(err)
	}

	if browser == nil {
		browser = webapp.setBrowserInCookie(w, r)
	}
	w.Write([]byte(html))

	webapp.logRequest(receivedAt, r, http.StatusOK, len(html), null.String{},
		browser)
}

func (webapp *WebApp) getStaticFile(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)

	bytes, err := ioutil.ReadFile(webapp.staticDir + r.URL.RequestURI())
	if os.IsNotExist(err) {
		webapp.notFound(w, r)
		return
	} else if err != nil {
		panic(err)
	}
	w.Write([]byte(bytes))

	webapp.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browser)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods",
		"DELETE, GET, PATCH, POST, PUT")
}

func (webapp *WebApp) notFound(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

	webapp.logRequest(receivedAt, r, http.StatusNotFound,
		len(http.StatusText(http.StatusNotFound)), null.String{}, browser)
}

func (webapp *WebApp) getWithoutTls(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)

	http.Redirect(w, r, "https://wellsaid.us"+r.RequestURI,
		http.StatusMovedPermanently)

	webapp.logRequest(receivedAt, r, http.StatusMovedPermanently, 0,
		null.String{}, browser)
}
