package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"bitbucket.org/danstutzman/wellsaid-backend/internal/model"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func (webapp *WebApp) postSync(r *http.Request,
	browser *db.BrowsersRow) Response {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var syncRequest model.SyncRequest
	err = json.Unmarshal(body, &syncRequest)
	if err != nil {
		panic(err)
	}

	webapp.model.PostSync(syncRequest, browser.Id)

	return JsonResponse{content: "OK"}
}
