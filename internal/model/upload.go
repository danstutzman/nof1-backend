package model

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"io"
	"os"
	"path"
	"strconv"
	"time"
)

type UploadRequest struct {
	Id         int     `json:"id"`
	Prompt     string  `json:"prompt"`
	RecordedAt float64 `json:"recordedAt"`
}

type UploadResponse struct {
	BackendUrl string `json:"backendUrl"`
}

func (model *Model) Upload(request UploadRequest,
	file io.Reader, userId int64, uploadedAt time.Time) UploadResponse {

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

	io.Copy(f, file)

	db.InsertIntoRecordings(model.dbConn, db.RecordingsRow{
		UserId:             userId,
		IdOnClient:         request.Id,
		RecordedAtOnClient: request.RecordedAt,
		UploadedAt:         uploadedAt,
		Path:               path,
		Prompt:             request.Prompt,
	})

	return UploadResponse{BackendUrl: "/recordings/" + filename}
}
