package webapp

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
	webapp      *WebApp
}

func NewRecoveryHandler(safeHandler http.Handler, webapp *WebApp) http.Handler {
	return &recoveryHandler{
		safeHandler: safeHandler,
		webapp:      webapp,
	}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()

	defer func() {
		if err := recover(); err != nil {
			browser := h.webapp.getBrowserFromCookie(r)

			fmt.Fprintln(os.Stderr, errors.Wrap(err, 2).ErrorStack())

			w.WriteHeader(http.StatusInternalServerError)

			h.webapp.logRequest(receivedAt, r, http.StatusInternalServerError, 0,
				null.StringFrom(errors.Wrap(err, 2).ErrorStack()), browser)
		}
	}()

	h.safeHandler.ServeHTTP(w, r)
}
