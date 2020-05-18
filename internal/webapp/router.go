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

	router.HandleFunc(
		"/admin/recordings",
		basicAuth(a.wrap(a.getAdminRecordings, false), "admin", a.adminPassword),
	).Methods("GET")

	router.HandleFunc(
		"/admin/recordings",
		basicAuth(a.wrap(a.postAdminRecordings, false), "admin", a.adminPassword),
	).Methods("POST")

	router.HandleFunc("/admin/recordings/{userId}/{filename}",
		basicAuth(a.wrap(a.getAdminRecording, false), "admin", a.adminPassword))

	return router
}

func NewRedirectToTlsRouter(a *WebApp) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(a.wrap(a.getWithoutTls, false))
	return router
}
