package app

import (
	"github.com/gorilla/mux"
	"net/http"
)

func NewRouter(app *App) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			app.NotFound(w, r)
		})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		app.GetRoot(w, r)
	})
	router.HandleFunc("/{prefix}.mp3",
		func(w http.ResponseWriter, r *http.Request) {
			app.GetStaticFile(w, r)
		})
	router.HandleFunc("/bundle.js",
		func(w http.ResponseWriter, r *http.Request) {
			app.GetStaticFile(w, r)
		})
	router.HandleFunc("/bundle.js.map",
		func(w http.ResponseWriter, r *http.Request) {
			app.GetStaticFile(w, r)
		})
	router.HandleFunc("/upload-audio",
		func(w http.ResponseWriter, r *http.Request) {
			app.PostUploadAudio(w, r)
		})
	router.HandleFunc("/sync",
		func(w http.ResponseWriter, r *http.Request) {
			app.PostSync(w, r)
		})
	return router
}

func NewRedirectToTlsRouter(app *App) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			app.GetWithoutTls(w, r)
		})
	return router
}
