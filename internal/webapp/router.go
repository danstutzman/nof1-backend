package webapp

import (
	"github.com/gorilla/mux"
	"net/http"
)

func NewRouter(webapp *WebApp) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			webapp.notFound(w, r)
		})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		webapp.getRoot(w, r)
	})
	router.HandleFunc("/{prefix}.mp3",
		func(w http.ResponseWriter, r *http.Request) {
			webapp.getStaticFile(w, r)
		})
	router.HandleFunc("/bundle.js",
		func(w http.ResponseWriter, r *http.Request) {
			webapp.getStaticFile(w, r)
		})
	router.HandleFunc("/bundle.js.map",
		func(w http.ResponseWriter, r *http.Request) {
			webapp.getStaticFile(w, r)
		})
	router.HandleFunc("/sync-with-audio",
		func(w http.ResponseWriter, r *http.Request) {
			webapp.postSyncWithAudio(w, r)
		})
	router.HandleFunc("/sync",
		func(w http.ResponseWriter, r *http.Request) {
			webapp.postSync(w, r)
		})
	return router
}

func NewRedirectToTlsRouter(webapp *WebApp) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			webapp.getWithoutTls(w, r)
		})
	return router
}
