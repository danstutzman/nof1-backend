package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"encoding/json"
	"io"
	"os"
	"path"
	"strconv"
	"time"
)

type UploadRequest struct {
	Id         int         `json:"id"`
	Metadata   interface{} `json:"metadata"`
	RecordedAt float64     `json:"recordedAt"`
}

type UploadResponse struct {
	BackendUrl string `json:"backendUrl"`
	Timestamp  string `json:"timestamp"`
}

func (model *Model) Upload(request UploadRequest, file io.Reader, userId int64,
	uploadedAt time.Time, mimeType string) UploadResponse {

	filename := strconv.FormatInt(uploadedAt.Unix(), 10)

	userDir := path.Join(model.uploadDir, strconv.FormatInt(userId, 10))
	err := os.MkdirAll(userDir, 0777)
	if err != nil {
		panic(err)
	}

	path := path.Join(userDir, filename)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	size, err := io.Copy(f, file)
	if err != nil {
		panic(err)
	}

	metadataJson, err := json.Marshal(request.Metadata)
	if err != nil {
		panic(err)
	}

	recording := db.InsertIntoRecordings(model.dbConn, db.RecordingsRow{
		UserId:             userId,
		IdOnClient:         request.Id,
		RecordedAtOnClient: request.RecordedAt,
		UploadedAt:         uploadedAt,
		Filename:           filename,
		MimeType:           mimeType,
		Size:               int(size),
		MetadataJson:       string(metadataJson),
	})

	go model.transcribeRecording(recording)

	return UploadResponse{
		BackendUrl: "/recordings/" + filename,
		Timestamp:  getTimestamp(),
	}
}
