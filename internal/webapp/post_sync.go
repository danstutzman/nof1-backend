package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/app"
	"encoding/json"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"net/http"
	"time"
)

func (webapp *WebApp) postSync(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browser := webapp.getBrowserFromCookie(r)
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=\"utf-8\"")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var syncRequest app.SyncRequest
	err = json.Unmarshal(body, &syncRequest)
	if err != nil {
		panic(err)
	}

	if browser == nil {
		browser = webapp.setBrowserInCookie(w, r)
	}
	webapp.app.PostSync(syncRequest, browser.Id)

	bytes := []byte("{}")
	w.Write(bytes)

	webapp.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browser)
}
