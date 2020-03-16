package db

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func generateToken() string {
	buffer := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, buffer)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}
