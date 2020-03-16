package webapp

import (
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"time"
)

func (webapp *WebApp) postUploadAudio(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := webapp.getBrowserTokenCookie(r)

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("audio_data")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.Printf("Header: %v", handler.Header)

	webapp.app.PostUploadAudio(file, handler.Filename)

	if browserId == 0 {
		browserId = webapp.setBrowserTokenCookie(w, r)
	}
	bytes := "OK"
	w.Write([]byte(bytes))

	webapp.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}
