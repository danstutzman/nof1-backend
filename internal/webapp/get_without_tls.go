package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"net/http"
)

func (webapp *WebApp) getWithoutTls(r *http.Request,
	browser *db.BrowsersRow) Response {

	return RedirectResponse{url: "https://n-of-1.club" + r.RequestURI}
}
