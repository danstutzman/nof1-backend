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
	secretKey   string
}

func newRecoveryHandler(safeHandler http.Handler, dbConn *sql.DB,
	secretKey string) http.Handler {
	return &recoveryHandler{
		safeHandler: safeHandler,
		dbConn:      dbConn,
		secretKey:   secretKey,
	}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()

	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, 2).ErrorStack())

			w.WriteHeader(http.StatusInternalServerError)

			browserId := getBrowserIdCookie(w, r, h.secretKey)

			logRequest(h.dbConn, receivedAt, r, http.StatusInternalServerError, 0,
				null.StringFrom(errors.Wrap(err, 2).ErrorStack()), browserId)
		}
	}()

	h.safeHandler.ServeHTTP(w, r)
}
