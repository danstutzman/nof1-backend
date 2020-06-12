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
		newTranscriptManual := r.FormValue(
			fmt.Sprintf("%d.transcriptManual", recording.Id))
		if newTranscriptManual != recording.TranscriptManual {
			recording.TranscriptManual = newTranscriptManual
			webapp.model.UpdateTranscriptManualOnRecording(recording)
		}
	}

	return RedirectResponse{url: "/admin/recordings"}
}
