package main

import (
	"database/sql"
	"fmt"
	"github.com/go-errors/errors"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"os"
	"time"
)

type recoveryHandler struct {
	safeHandler http.Handler
	dbConn      *sql.DB
}

func newRecoveryHandler(safeHandler http.Handler, dbConn *sql.DB) http.Handler {
	return &recoveryHandler{safeHandler: safeHandler, dbConn: dbConn}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := getBrowserIdCookie(w, r)

	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, 2).ErrorStack())

			w.WriteHeader(http.StatusInternalServerError)

			logRequest(h.dbConn, receivedAt, r, http.StatusInternalServerError, 0,
				null.StringFrom(errors.Wrap(err, 2).ErrorStack()), browserId)
		}
	}()

	h.safeHandler.ServeHTTP(w, r)
}
