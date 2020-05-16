package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"fmt"
	"path"
	"strconv"
)

type GetRecordingResponse struct {
	Path     string
	MimeType string
	Size     int
}

func (model *Model) GetRecording(userId int64,
	filename string) *GetRecordingResponse {

	recordings := db.FromRecordings(model.dbConn,
		fmt.Sprintf("WHERE user_id=%d AND filename=%s",
			userId, db.EscapeString(filename)))
	if len(recordings) == 0 {
		return nil
	}

	userDir := path.Join(model.uploadDir, strconv.FormatInt(userId, 10))
	audioPath := path.Join(userDir, recordings[0].Filename)

	return &GetRecordingResponse{
		Path:     audioPath,
		MimeType: recordings[0].MimeType,
		Size:     recordings[0].Size,
	}
}
