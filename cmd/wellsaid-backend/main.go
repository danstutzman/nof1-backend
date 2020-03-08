package main

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"crypto/tls"
	"database/sql"
	"github.com/gorilla/mux"
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"os"
	"time"
)

func logRequest(dbConn *sql.DB, receivedAt time.Time, r *http.Request,
	statusCode, size int) db.RequestsRow {

	var tlsProtocol null.String
	var tlsCipher null.String
	if r.TLS != nil {
		tlsProtocol = null.StringFrom(r.TLS.NegotiatedProtocol)
		tlsCipher = null.StringFrom(tls.CipherSuiteName(r.TLS.CipherSuite))
	}

	return db.InsertIntoRequests(dbConn, db.RequestsRow{
		ReceivedAt:  receivedAt,
		RemoteAddr:  r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Referer:     r.Referer(),
		HttpVersion: r.Proto,
		TlsProtocol: tlsProtocol,
		TlsCipher:   tlsCipher,
		Method:      r.Method,
		Path:        r.URL.RequestURI(),
		DurationMs:  int(time.Now().UTC().Sub(receivedAt).Milliseconds()),
		StatusCode:  statusCode,
		Size:        size,
	})
}

func getRoot(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

	html := `<!DOCTYPE html>
<html lang='en'>
  <head>
    <meta charset='utf-8'>
    <title>Well Said</title>
  </head>
  <body>
		Loaded.
	  <script>
		  var xhr = new XMLHttpRequest()
      xhr.open('POST', '/capabilities', true)
			xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded')
      xhr.send('a=b' +
				'&nacn=' + encodeURIComponent(navigator.appCodeName) +
				'&nan=' + encodeURIComponent(navigator.appName) +
				'&nav=' + encodeURIComponent(navigator.appVersion) +
				'&nce=' + encodeURIComponent(navigator.cookieEnabled) +
				'&nl=' + encodeURIComponent(navigator.language) +
				'&nls=' + encodeURIComponent(navigator.languages) +
				'&np=' + encodeURIComponent(navigator.platform) +
				'&no=' + encodeURIComponent(navigator.oscpu) +
				'&nua=' + encodeURIComponent(navigator.userAgent) +
				'&nv=' + encodeURIComponent(navigator.vendor) +
				'&nvs=' + encodeURIComponent(navigator.vendorSub) +
				'&sw=' + screen.width +
				'&sh=' + screen.height +
				'&wiw=' + window.innerWidth +
        '&wih=' + window.innerHeight +
				'&dbcw=' + document.body.clientWidth +
				'&dbch=' + document.body.clientHeight +
				'&ddecw=' + document.documentElement.clientWidth +
				'&ddech=' + document.documentElement.clientHeight +
				'&wsaw=' + window.screen.availWidth +
				'&wsah=' + window.screen.availHeight +
				'&wdpr=' + window.devicePixelRatio +
				'&ddeots=' + ('ontouchstart' in document.documentElement)
			)
		</script>
  </body>
</html>
`
	w.Write([]byte(html))

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(html))
}

func param(r *http.Request, key string) null.String {
	values := r.Form[key]
	if len(values) == 1 {
		return null.StringFrom(values[0])
	} else {
		return null.String{}
	}
}

func postCapabilities(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

	w.Write([]byte("OK"))

	requestLog := logRequest(dbConn, receivedAt, r, http.StatusOK, len("OK"))

	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	db.InsertIntoCapabilities(dbConn, db.CapabilitiesRow{
		RequestId:               requestLog.Id,
		NavigatorAppCodeName:    param(r, "nacn"),
		NavigatorAppName:        param(r, "nan"),
		NavigatorAppVersion:     param(r, "nav"),
		NavigatorCookieEnabled:  param(r, "nce"),
		NavigatorLanguage:       param(r, "nl"),
		NavigatorLanguages:      param(r, "nls"),
		NavigatorPlatform:       param(r, "np"),
		NavigatorOscpu:          param(r, "no"),
		NavigatorUserAgent:      param(r, "nua"),
		NavigatorVendor:         param(r, "nv"),
		NavigatorVendorSub:      param(r, "nvs"),
		ScreenWidth:             param(r, "sw"),
		ScreenHeight:            param(r, "sh"),
		WindowInnerWidth:        param(r, "wiw"),
		WindowInnerHeight:       param(r, "wih"),
		DocBodyClientWidth:      param(r, "dbcw"),
		DocBodyClientHeight:     param(r, "dbch"),
		DocElementClientWidth:   param(r, "ddecw"),
		DocElementClientHeight:  param(r, "ddech"),
		WindowScreenAvailWidth:  param(r, "wsaw"),
		WindowScreenAvailHeight: param(r, "wsah"),
		WindowDevicePixelRatio:  param(r, "wdpr"),
		HasOnTouchStart:         param(r, "ddeots"),
	})
}

func notFound(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

	output := "Not found\n"
	w.Write([]byte(output))

	logRequest(dbConn, receivedAt, r, http.StatusNotFound, len(output))
}

func getWithoutTls(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

	http.Redirect(w, r, "https://wellsaid.us"+r.RequestURI,
		http.StatusMovedPermanently)

	logRequest(dbConn, receivedAt, r, http.StatusMovedPermanently, 0)
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		log.Fatalf("Set HTTP_PORT env var")
	}

	httpsCertFile := os.Getenv("HTTPS_CERT_FILE")
	httpsKeyFile := os.Getenv("HTTPS_KEY_FILE")

	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		log.Fatalf("Set DB_FILE env var.")
	}
	dbConn := db.InitDb(dbFile)

	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			notFound(w, r, dbConn)
		})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		getRoot(w, r, dbConn)
	})
	router.HandleFunc("/capabilities",
		func(w http.ResponseWriter, r *http.Request) {
			postCapabilities(w, r, dbConn)
		})

	if httpsCertFile != "" || httpsKeyFile != "" {
		log.Printf("Serving TLS on :443 and HTTP on :" + httpPort + "...")

		go func() {
			err := http.ListenAndServeTLS(":443", httpsCertFile, httpsKeyFile, router)
			panic(err)
		}()

		redirectToTlsRouter := mux.NewRouter()
		redirectToTlsRouter.NotFoundHandler = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				getWithoutTls(w, r, dbConn)
			})
		err := http.ListenAndServe(":"+httpPort, redirectToTlsRouter)
		panic(err)
	} else {
		log.Printf("Serving HTTP on :" + httpPort + "...")
		err := http.ListenAndServe(":"+httpPort, router)
		panic(err)
	}
}
