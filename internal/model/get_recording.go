package model

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

func (model *Model) GetRecording(userId int64,
	filename string) ([]byte, error) {

	userDir := path.Join(model.uploadDir, strconv.FormatInt(userId, 10))
	err := os.MkdirAll(userDir, 0777)
	if err != nil {
		panic(err)
	}

	audioPath := path.Join(userDir, filename)

	bytes, err := ioutil.ReadFile(audioPath)
	return bytes, err
}
