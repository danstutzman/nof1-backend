package webapp

import (
	"github.com/gorilla/mux"
	"gopkg.in/guregu/null.v3"
	"net/http"
	"os"
	"time"
)

func (webapp *WebApp) getRecording(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)
	params := mux.Vars(r)

	var bytes []byte
	var status int
	if !browser.UserId.Valid {
		bytes = []byte("Unauthorized")
		status = http.StatusUnauthorized
		http.Error(w, "Unauthorized", status)
	} else {
		bytes, err := webapp.model.GetRecording(
			browser.UserId.Int64, params["filename"])
		if os.IsNotExist(err) {
			bytes = []byte("Not found")
			status = http.StatusNotFound
			http.Error(w, "Not found", status)
		} else if err != nil {
			panic(err)
		} else {
			status = http.StatusOK
			w.Write(bytes)
		}
	}

	webapp.logRequest(receivedAt, r, status, len(bytes), null.String{},
		browser)
}
