package main

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"crypto/tls"
	"database/sql"
	"github.com/gorilla/mux"
	"gopkg.in/guregu/null.v3"
	"io"
	"io/ioutil"
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

func getRoot(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	staticDir string) {
	receivedAt := time.Now().UTC()

	html, err := ioutil.ReadFile(staticDir + "/index.html")
	if err != nil {
		panic(err)
	}
	w.Write([]byte(html))

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(html))
}

func getStaticFile(w http.ResponseWriter, r *http.Request, dbConn *sql.DB,
	staticDir string) {
	receivedAt := time.Now().UTC()

	bytes, err := ioutil.ReadFile(staticDir + r.URL.RequestURI())
	if err != nil {
		panic(err)
	}
	w.Write([]byte(bytes))

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(bytes))
}

func postUploadAudio(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	receivedAt := time.Now().UTC()

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

	logRequest(dbConn, receivedAt, r, http.StatusOK, len(bytes))
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
			notFound(w, r, dbConn)
		})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		getRoot(w, r, dbConn, staticDir)
	})
	router.HandleFunc("/{prefix}.mp3",
		func(w http.ResponseWriter, r *http.Request) {
			getStaticFile(w, r, dbConn, staticDir)
		})
	router.HandleFunc("/capabilities",
		func(w http.ResponseWriter, r *http.Request) {
			postCapabilities(w, r, dbConn)
		})
	router.HandleFunc("/bundle.js",
		func(w http.ResponseWriter, r *http.Request) {
			getStaticFile(w, r, dbConn, staticDir)
		})
	router.HandleFunc("/bundle.js.map",
		func(w http.ResponseWriter, r *http.Request) {
			getStaticFile(w, r, dbConn, staticDir)
		})
	router.HandleFunc("/upload-audio",
		func(w http.ResponseWriter, r *http.Request) {
			postUploadAudio(w, r, dbConn)
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
