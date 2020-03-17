package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"time"
)

func (webapp *WebApp) postUploadAudio(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("audio_data")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log.Printf("Header: %v", handler.Header)

	if browser == nil {
		browser = webapp.setBrowserInCookie(w, r)
	}
	if !browser.UserId.Valid {
		userId := db.InsertIntoUsers(webapp.dbConn)
		browser.UserId = null.IntFrom(userId)
	}
	webapp.model.PostUploadAudio(file, browser.UserId.Int64, handler.Filename)

	bytes := "OK"
	w.Write([]byte(bytes))

	db.UpdateUserIdAndLastSeenAtOnBrowser(
		webapp.dbConn, browser.UserId.Int64, browser.Id)
	webapp.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browser)
}
