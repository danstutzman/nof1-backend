package webapp

import (
	"github.com/gorilla/mux"
	"net/http"
)

func NewRouter(a *WebApp) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(a.wrap(a.notFound, false))
	router.HandleFunc("/", a.wrap(a.getRoot, true))
	router.HandleFunc("/record/{metadata}", a.wrap(a.getRoot, true))
	router.HandleFunc("/{prefix}.svg", a.wrap(a.getStaticFile, false))
	router.HandleFunc("/bundle.js", a.wrap(a.getStaticFile, false))
	router.HandleFunc("/bundle.js.map", a.wrap(a.getStaticFile, false))
	router.HandleFunc("/upload", a.wrap(a.postUpload, true))
	router.HandleFunc("/sync", a.wrap(a.postSync, true))
	router.HandleFunc("/recordings/{filename}", a.wrap(a.getRecording, false))
	return router
}

func NewRedirectToTlsRouter(a *WebApp) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(a.wrap(a.getWithoutTls, false))
	return router
}
