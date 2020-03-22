package webapp

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"net/http"
)

func (webapp *WebApp) getWithoutTls(r *http.Request,
	browser *db.BrowsersRow) Response {

	return RedirectResponse{url: "https://wellsaid.us" + r.RequestURI}
}
