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
	Status(responseHeaders http.Header) int
	Size() int
	ServeHTTP(w http.ResponseWriter, r *http.Request) http.Header
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

func (response JsonResponse) Status(responseHeaders http.Header) int {
	return http.StatusOK
}

func (response JsonResponse) Size() int {
	return len(response.getContentBytes())
}

func (response JsonResponse) ServeHTTP(w http.ResponseWriter,
	r *http.Request) http.Header {

	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=\"utf-8\"")
	w.Write(response.getContentBytes())
	return w.Header()
}

type BytesResponse struct {
	content     []byte
	contentType string
}

func (response BytesResponse) Status(responseHeaders http.Header) int {
	return http.StatusOK
}
func (response BytesResponse) Size() int { return len(response.content) }
func (response BytesResponse) ServeHTTP(w http.ResponseWriter,
	r *http.Request) http.Header {

	w.Header().Set("Content-Type", response.contentType)
	w.Write(response.content)
	return w.Header()
}

type ErrorResponse struct {
	status int
}

func (response ErrorResponse) Status(responseHeaders http.Header) int {
	return response.status
}
func (response ErrorResponse) Size() int { return len(response.getContent()) }
func (response ErrorResponse) getContent() []byte {
	return []byte(http.StatusText(response.status))
}
func (response ErrorResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) http.Header {
	http.Error(w, string(response.getContent()), response.status)
	return w.Header()
}

type BadRequestResponse struct {
	message string
}

func (response BadRequestResponse) Status(responseHeaders http.Header) int {
	return http.StatusBadRequest
}
func (response BadRequestResponse) Size() int {
	return len([]byte(response.message))
}

func (response BadRequestResponse) ServeHTTP(w http.ResponseWriter,
	r *http.Request) http.Header {
	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, response.message, http.StatusBadRequest)
	return w.Header()
}

type RedirectResponse struct {
	url string
}

func (response RedirectResponse) Status(responseHeaders http.Header) int {
	return http.StatusMovedPermanently
}
func (response RedirectResponse) Size() int {
	return len(response.getContent())
}
func (response RedirectResponse) getContent() []byte {
	return []byte("<a href=\"" + response.url + "\">" + "Please click here to be redirected." + "</a>.\n")
}
func (response RedirectResponse) ServeHTTP(w http.ResponseWriter,
	r *http.Request) http.Header {
	h := w.Header()

	h.Set("Location", response.url)

	// RFC 7231 notes that a short HTML body is usually included in
	// the response because older user agents may not understand 301/307.
	h.Set("Content-Type", "text/html; charset=utf-8")

	http.Error(w, string(response.getContent()), http.StatusMovedPermanently)

	return w.Header()
}

type FileResponse struct {
	path          string
	size          int
	mimeType      string
	servedHeaders http.Header
}

func (response FileResponse) Status(responseHeaders http.Header) int {
	if len(responseHeaders["Last-Modified"]) == 0 {
		panic("FileResponse's Status() can't be called until after ServeHTTP")
	} else if len(responseHeaders["Content-Length"]) == 0 {
		return http.StatusNotModified
	} else {
		return http.StatusOK
	}
}
func (response FileResponse) Size() int {
	return response.size
}
func (response FileResponse) ServeHTTP(w http.ResponseWriter,
	r *http.Request) http.Header {
	if response.mimeType != "" {
		w.Header().Set("Content-Type", response.mimeType)
	}
	http.ServeFile(w, r, response.path)
	return w.Header()
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

		responseHeaders := response.ServeHTTP(w, r)

		webapp.logRequest(receivedAt, r, response.Status(responseHeaders),
			response.Size(), null.String{}, browser)
	}
}
