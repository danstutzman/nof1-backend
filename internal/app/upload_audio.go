package app

import (
	"io"
	"os"
)

func (app *App) PostUploadAudio(file io.Reader, filename string) {
	path := app.uploadDir + filename

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	io.Copy(f, file)
}
