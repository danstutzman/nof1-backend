package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
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
	Content() []byte
	SetHeaders(w http.ResponseWriter)
}

type JsonResponse struct {
	content interface{}
}

func (response JsonResponse) Status() int { return http.StatusOK }

func (response JsonResponse) Content() []byte {
	bytes, err := json.Marshal(response.content)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (response JsonResponse) SetHeaders(w http.ResponseWriter) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=\"utf-8\"")
}

type BytesResponse struct {
	content     []byte
	contentType string
}

func (response BytesResponse) Status() int { return http.StatusOK }

func (response BytesResponse) Content() []byte { return response.content }

func (response BytesResponse) SetHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", response.contentType)
}

type ErrorResponse struct {
	status int
}

func (response ErrorResponse) Status() int { return response.status }

func (response ErrorResponse) Content() []byte {
	return []byte(http.StatusText(response.status))
}

func (response ErrorResponse) SetHeaders(w http.ResponseWriter) {}

type BadRequestResponse struct {
	message string
}

func (response BadRequestResponse) Status() int { return http.StatusBadRequest }
func (response BadRequestResponse) Content() []byte {
	return []byte(response.message)
}
func (response BadRequestResponse) SetHeaders(w http.ResponseWriter) {}

type RedirectResponse struct {
	url string
}

func (response RedirectResponse) Status() int {
	return http.StatusMovedPermanently
}
func (response RedirectResponse) Content() []byte {
	return []byte("<a href=\"" + response.url + "\">" + "Redirect" + "</a>.\n")
}
func (response RedirectResponse) SetHeaders(w http.ResponseWriter) {
	h := w.Header()

	h.Set("Location", response.url)

	// RFC 7231 notes that a short HTML body is usually included in
	// the response because older user agents may not understand 301/307.
	h.Set("Content-Type", "text/html; charset=utf-8")
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

		content := response.Content()
		response.SetHeaders(w)
		w.Write(content)

		webapp.logRequest(receivedAt, r, response.Status(), len(content),
			null.String{}, browser)
	}
}
