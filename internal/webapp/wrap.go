package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"os"
	"time"
)

type Response interface {
	Status() int
	Size() int
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type JsonResponse struct {
	content      interface{}
	contentBytes []byte
}

func (response JsonResponse) getContentBytes() []byte {
	if len(response.contentBytes) == 0 {
		var err error
		response.contentBytes, err = json.Marshal(response.content)
		if err != nil {
			panic(err)
		}
	}
	return response.contentBytes
}

func (response JsonResponse) Status() int { return http.StatusOK }

func (response JsonResponse) Size() int {
	return len(response.getContentBytes())
}

func (response JsonResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=\"utf-8\"")
	w.Write(response.getContentBytes())
}

type BytesResponse struct {
	content     []byte
	contentType string
}

func (response BytesResponse) Status() int { return http.StatusOK }

func (response BytesResponse) Size() int { return len(response.content) }

func (response BytesResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", response.contentType)
	w.Write(response.content)
}

type ErrorResponse struct {
	status int
}

func (response ErrorResponse) Status() int { return response.status }

func (response ErrorResponse) Size() int { return len(response.getContent()) }

func (response ErrorResponse) getContent() []byte {
	return []byte(http.StatusText(response.status))
}

func (response ErrorResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, string(response.getContent()), response.status)
}

type BadRequestResponse struct {
	message string
}

func (response BadRequestResponse) Status() int { return http.StatusBadRequest }
func (response BadRequestResponse) Size() int {
	return len([]byte(response.message))
}

func (response BadRequestResponse) ServeHTTP(w http.ResponseWriter,
	r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(response.message))
}

type RedirectResponse struct {
	url string
}

func (response RedirectResponse) Status() int {
	return http.StatusMovedPermanently
}
func (response RedirectResponse) Size() int {
	return len(response.getContent())
}
func (response RedirectResponse) getContent() []byte {
	return []byte("<a href=\"" + response.url + "\">" + "Please click here to be redirected." + "</a>.\n")
}
func (response RedirectResponse) ServeHTTP(w http.ResponseWriter,
	r *http.Request) {
	h := w.Header()

	h.Set("Location", response.url)

	// RFC 7231 notes that a short HTML body is usually included in
	// the response because older user agents may not understand 301/307.
	h.Set("Content-Type", "text/html; charset=utf-8")

	http.Error(w, string(response.getContent()), response.Status())
}

type FileResponse struct {
	path     string
	size     int
	mimeType string
}

func (response FileResponse) Status() int {
	return http.StatusOK
}
func (response FileResponse) Size() int {
	return response.size
}
func (response FileResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if response.mimeType != "" {
		w.Header().Set("Content-Type", response.mimeType)
	}
	http.ServeFile(w, r, response.path)
}

type HandlerFunc func(r *http.Request, browser *db.BrowsersRow) Response

func (webapp *WebApp) wrap(handler HandlerFunc,
	setBrowserCookie bool) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		receivedAt := time.Now().UTC()
		browser := webapp.getBrowserFromCookie(r)

		defer func() {
			if err := recover(); err != nil {
				errorStack := errors.Wrap(err, 2).ErrorStack()

				fmt.Fprintln(os.Stderr, errorStack)

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

				webapp.logRequest(receivedAt, r, http.StatusInternalServerError,
					len([]byte(errorStack)), null.StringFrom(errorStack), browser)
			}
		}()

		if browser == nil && setBrowserCookie {
			browser = webapp.setBrowserInCookie(w, r)
		}

		response := handler(r, browser)

		response.ServeHTTP(w, r)

		webapp.logRequest(receivedAt, r, response.Status(), response.Size(),
			null.String{}, browser)
	}
}
