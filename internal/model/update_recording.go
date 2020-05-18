package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
)

func (model *Model) UpdateTranscriptOnRecording(recording db.RecordingsRow) {
	db.UpdateTranscriptOnRecording(model.dbConn,
		recording.Transcript, recording.Id)
}
