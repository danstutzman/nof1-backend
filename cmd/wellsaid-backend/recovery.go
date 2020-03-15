package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"os"
	"time"
)

type recoveryHandler struct {
	safeHandler http.Handler
	app         App
}

func newRecoveryHandler(safeHandler http.Handler, app App) http.Handler {
	return &recoveryHandler{
		safeHandler: safeHandler,
		app:         app,
	}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()

	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, 2).ErrorStack())

			w.WriteHeader(http.StatusInternalServerError)

			browserId := h.app.getBrowserIdCookie(w, r)

			h.app.logRequest(receivedAt, r, http.StatusInternalServerError, 0,
				null.StringFrom(errors.Wrap(err, 2).ErrorStack()), browserId)
		}
	}()

	h.safeHandler.ServeHTTP(w, r)
}
