package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
)

func (model *Model) UpdateTranscriptManualOnRecording(
	recording db.RecordingsRow) {
	db.UpdateTranscriptManualOnRecording(model.dbConn,
		recording.TranscriptManual, recording.Id)
}
