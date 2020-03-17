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
	file, handler, fileErr := r.FormFile("audio_data")
	if fileErr == nil {
		defer file.Close()
		log.Printf("Header: %v", handler.Header)
	}

	if browser == nil {
		browser = webapp.setBrowserInCookie(w, r)
	}
	if !browser.UserId.Valid {
		userId := db.InsertIntoUsers(webapp.dbConn)
		browser.UserId = null.IntFrom(userId)
	}

	var bytes string
	var status int
	if fileErr == nil {
		webapp.model.PostUploadAudio(file, browser.UserId.Int64, handler.Filename)
		bytes = "OK"
		status = http.StatusOK
	} else if fileErr == http.ErrMissingFile {
		bytes = "Missing file"
		status = http.StatusBadRequest
	} else if fileErr == http.ErrMissingBoundary {
		bytes = fileErr.Error()
		status = http.StatusBadRequest
	} else if fileErr == http.ErrNotMultipart {
		bytes = fileErr.Error()
		status = http.StatusBadRequest
	} else {
		bytes = "Internal server error"
		status = http.StatusInternalServerError
	}
	http.Error(w, bytes, status)

	db.UpdateUserIdAndLastSeenAtOnBrowser(
		webapp.dbConn, browser.UserId.Int64, browser.Id)
	webapp.logRequest(receivedAt, r, status, len(bytes), null.String{}, browser)
}
