package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"bitbucket.org/danstutzman/nof1-backend/internal/model"
	"encoding/json"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"time"
)

func (webapp *WebApp) postUpload(r *http.Request,
	browser *db.BrowsersRow) Response {

	if !browser.UserId.Valid {
		userId := db.InsertIntoUsers(webapp.dbConn)
		browser.UserId = null.IntFrom(userId)
	}

	r.ParseMultipartForm(32 << 20)

	var syncRequest model.SyncRequest
	err := json.Unmarshal([]byte(r.FormValue("sync_request")), &syncRequest)
	if err != nil {
		return BadRequestResponse{message: err.Error()}
	}

	webapp.model.PostSync(syncRequest, browser.Id)

	file, fileHeader, err := r.FormFile("audio_data")
	if err == http.ErrMissingFile ||
		err == http.ErrMissingBoundary ||
		err == http.ErrNotMultipart {
		return BadRequestResponse{message: "Bad audio_data param: " + err.Error()}
	}
	defer file.Close()

	var request model.UploadRequest
	err = json.Unmarshal([]byte(r.FormValue("recording")), &request)
	if err != nil {
		return BadRequestResponse{message: "Bad recording param: " + err.Error()}
	}
	content := webapp.model.Upload(request, file, browser.UserId.Int64,
		time.Now().UTC(), fileHeader.Header.Get("Content-Type"))

	db.UpdateUserIdAndLastSeenAtOnBrowser(
		webapp.dbConn, browser.UserId.Int64, browser.Id)

	return JsonResponse{content: content}
}
