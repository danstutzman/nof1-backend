package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/model"
	"encoding/json"
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
)

func (webapp *WebApp) postSyncWithAudio(r *http.Request,
	browser *db.BrowsersRow) Response {

	if !browser.UserId.Valid {
		userId := db.InsertIntoUsers(webapp.dbConn)
		browser.UserId = null.IntFrom(userId)
	}

	var syncRequest model.SyncRequest
	err := json.Unmarshal([]byte(r.FormValue("sync_request")), &syncRequest)
	if err != nil {
		return BadRequestResponse{message: err.Error()}
	}

	webapp.model.PostSync(syncRequest, browser.Id)

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("audio_data")
	if err == http.ErrMissingFile ||
		err == http.ErrMissingBoundary ||
		err == http.ErrNotMultipart {
		return BadRequestResponse{message: err.Error()}
	}
	defer file.Close()
	log.Printf("Header: %v", handler.Header)

	upload := webapp.model.UploadAudio(file, browser.UserId.Int64)

	db.UpdateUserIdAndLastSeenAtOnBrowser(
		webapp.dbConn, browser.UserId.Int64, browser.Id)

	return JsonResponse{content: upload}
}
