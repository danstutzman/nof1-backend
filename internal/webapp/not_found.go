package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"net/http"
)

func (webapp *WebApp) notFound(r *http.Request,
	browser *db.BrowsersRow) Response {

	return ErrorResponse{status: http.StatusNotFound}
}
