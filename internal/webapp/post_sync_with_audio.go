package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/model"
	"encoding/json"
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"time"
)

func (webapp *WebApp) postSyncWithAudio(w http.ResponseWriter,
	r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)

	r.ParseMultipartForm(32 << 20)
	file, handler, fileErr := r.FormFile("audio_data")
	if fileErr == nil {
		defer file.Close()
		log.Printf("Header: %v", handler.Header)
	}

	var syncRequest model.SyncRequest
	err := json.Unmarshal([]byte(r.FormValue("sync_request")), &syncRequest)
	if err != nil {
		panic(err)
	}

	if browser == nil {
		browser = webapp.setBrowserInCookie(w, r)
	}
	if !browser.UserId.Valid {
		userId := db.InsertIntoUsers(webapp.dbConn)
		browser.UserId = null.IntFrom(userId)
	}

	webapp.model.PostSync(syncRequest, browser.Id)

	var text string
	var status int
	if fileErr == nil {
		upload := webapp.model.UploadAudio(file, browser.UserId.Int64)

		bytes, err := json.Marshal(upload)
		if err != nil {
			panic(err)
		}
		text = string(bytes)
		status = http.StatusOK
	} else if fileErr == http.ErrMissingFile {
		text = "Missing file"
		status = http.StatusBadRequest
	} else if fileErr == http.ErrMissingBoundary {
		text = fileErr.Error()
		status = http.StatusBadRequest
	} else if fileErr == http.ErrNotMultipart {
		text = fileErr.Error()
		status = http.StatusBadRequest
	} else {
		text = "Internal server error"
		status = http.StatusInternalServerError
	}
	http.Error(w, text, status)

	db.UpdateUserIdAndLastSeenAtOnBrowser(
		webapp.dbConn, browser.UserId.Int64, browser.Id)
	webapp.logRequest(receivedAt, r, status, len(text), null.String{}, browser)
}
