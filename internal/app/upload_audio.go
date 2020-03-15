package app

import (
	"gopkg.in/guregu/null.v3"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func (app *App) PostUploadAudio(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getBrowserIdCookie(w, r)

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

	app.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}
