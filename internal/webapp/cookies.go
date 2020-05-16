package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"fmt"
	"log"
	"net/http"
	"time"
)

const COOKIE_NAME = "browser-token"

func (webapp *WebApp) getBrowserFromCookie(r *http.Request) *db.BrowsersRow {
	cookie, err := r.Cookie(COOKIE_NAME)
	if err == nil {
		browsers := db.FromBrowsers(webapp.dbConn,
			fmt.Sprintf("WHERE token=%s LIMIT 1", db.EscapeString(cookie.Value)))
		if len(browsers) == 0 {
			log.Printf("No browser row for %s cookie", COOKIE_NAME)
			return nil
		}
		return &browsers[0]
	} else if err == http.ErrNoCookie {
		return nil
	} else {
		panic(err)
	}
}

func (webapp *WebApp) setBrowserInCookie(w http.ResponseWriter,
	r *http.Request) *db.BrowsersRow {

	browser := db.InsertIntoBrowsers(webapp.dbConn, db.BrowsersRow{
		UserAgent:      r.UserAgent(),
		Accept:         r.Header.Get("Accept"),
		AcceptEncoding: r.Header.Get("Accept-Encoding"),
		AcceptLanguage: r.Header.Get("Accept-Language"),
		Referer:        r.Referer(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:    COOKIE_NAME,
		Value:   browser.Token,
		Expires: time.Now().AddDate(30, 0, 0),
	})

	return &browser
}
