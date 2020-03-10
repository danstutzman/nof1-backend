package main

import (
	"database/sql"
	"github.com/go-errors/errors"
	"gopkg.in/guregu/null.v3"
	"net/http"
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

	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logRequest(h.dbConn, receivedAt, r, http.StatusInternalServerError, 0,
				null.StringFrom(errors.Wrap(err, 2).ErrorStack()))
		}
	}()

	h.safeHandler.ServeHTTP(w, r)
}
