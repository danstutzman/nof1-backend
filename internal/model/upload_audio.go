package model

import (
	"io"
	"os"
	"path"
	"strconv"
)

func (model *Model) PostUploadAudio(file io.Reader, userId int64,
	filename string) {

	userDir := path.Join(model.uploadDir, strconv.FormatInt(userId, 10))
	err := os.MkdirAll(userDir, 0777)
	if err != nil {
		panic(err)
	}

	audioPath := path.Join(userDir, filename)

	f, err := os.OpenFile(audioPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	io.Copy(f, file)
}
