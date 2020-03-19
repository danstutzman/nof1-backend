package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func (webapp *WebApp) getRecording(r *http.Request,
	browser *db.BrowsersRow) Response {

	if !browser.UserId.Valid {
		return ErrorResponse{status: http.StatusUnauthorized}
	}

	params := mux.Vars(r)
	filename := params["filename"]
	if filename == "" {
		return BadRequestResponse{message: "Supply filename param"}
	}

	bytes, err := webapp.model.GetRecording(browser.UserId.Int64, filename)

	if os.IsNotExist(err) {
		return ErrorResponse{status: http.StatusNotFound}
	} else if err != nil {
		panic(err)
	}

	return BytesResponse{content: bytes}
}
