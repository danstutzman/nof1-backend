package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"github.com/gorilla/mux"
	"net/http"
)

func (webapp *WebApp) getRecording(r *http.Request,
	browser *db.BrowsersRow) Response {

	if browser == nil || !browser.UserId.Valid {
		return ErrorResponse{status: http.StatusUnauthorized}
	}

	params := mux.Vars(r)
	filename := params["filename"]
	if filename == "" {
		return BadRequestResponse{message: "Supply filename param"}
	}

	recording := webapp.model.GetRecording(browser.UserId.Int64, filename)
	if recording == nil {
		return ErrorResponse{status: http.StatusNotFound}
	}

	return FileResponse{
		path:     recording.Path,
		size:     recording.Size,
		mimeType: recording.MimeType,
	}
}
