package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"fmt"
	"net/http"
)

func (webapp *WebApp) postAdminRecordings(r *http.Request,
	browser *db.BrowsersRow) Response {

	err := r.ParseForm()
	if err != nil {
		return BadRequestResponse{message: "Couldn't parse form"}
	}

	for _, recording := range webapp.model.GetRecordings() {
		newTranscript := r.FormValue(fmt.Sprintf("%d.transcript", recording.Id))
		if newTranscript != recording.Transcript {
			recording.Transcript = newTranscript
			webapp.model.UpdateTranscriptOnRecording(recording)
		}
	}

	return RedirectResponse{url: "/admin/recordings"}
}
