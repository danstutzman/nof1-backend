package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"io"
)

func encrypt(plaintext string, secretKeyBase64 string) string {
	secretKey, err := base64.StdEncoding.DecodeString(secretKeyBase64)
	if err != nil {
		panic(err)
	}
	if len(secretKey) != 32 {
		panic("Expected 32 bytes Base64ed")
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		panic(err)
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	h := hmac.New(sha256.New, secretKey)
	h.Write(nonce)
	h.Write(ciphertext)
	mac := h.Sum(nil)

	concatted := append(append(nonce, []byte(ciphertext)...), []byte(mac)...)
	return base64.StdEncoding.EncodeToString(concatted)
}

func decrypt(noncePlusCipherTextBase64 string, secretKeyBase64 string) (string,
	error) {
	secretKey, err := base64.StdEncoding.DecodeString(secretKeyBase64)
	if err != nil {
		panic(err)
	}
	if len(secretKey) != 32 {
		panic("Expected 32 bytes Base64ed")
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	concatted, err := base64.StdEncoding.DecodeString(noncePlusCipherTextBase64)
	if err != nil {
		return "", err
	}

	if len(concatted) < gcm.NonceSize()+32 {
		return "", errors.New("Too short")
	}
	nonce, ciphertext, suppliedMac := concatted[:gcm.NonceSize()],
		concatted[gcm.NonceSize():len(concatted)-32],
		concatted[len(concatted)-32:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, secretKey)
	h.Write(nonce)
	h.Write(ciphertext)
	expectedMac := h.Sum(nil)
	if len(suppliedMac) != len(expectedMac) ||
		subtle.ConstantTimeCompare(suppliedMac, expectedMac) != 1 {
		return "", errors.New("Unexpected MAC")
	}

	return string(plaintext), nil
}

func IsSecretKeyOkay(secretKeyBase64 string) bool {
	secretKey, err := base64.StdEncoding.DecodeString(secretKeyBase64)
	return err == nil && len(secretKey) == 32
}

func MakeExampleSecretKey() string {
	example := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, example)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(example)
}
