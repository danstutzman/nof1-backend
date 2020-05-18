package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (webapp *WebApp) getAdminRecording(r *http.Request,
	browser *db.BrowsersRow) Response {

	params := mux.Vars(r)
	userIdString := params["userId"]
	filename := params["filename"]
	if userIdString == "" {
		return BadRequestResponse{message: "Supply userId param"}
	}
	if filename == "" {
		return BadRequestResponse{message: "Supply filename param"}
	}
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		return BadRequestResponse{message: "userId param must be an int"}
	}

	recording := webapp.model.GetRecording(int64(userId), filename)
	if recording == nil {
		return ErrorResponse{status: http.StatusNotFound}
	}

	return FileResponse{
		path:     recording.Path,
		size:     recording.Size,
		mimeType: recording.MimeType,
	}
}
