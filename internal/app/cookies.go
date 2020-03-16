package app

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"log"
	"net/http"
	"time"
)

const COOKIE_NAME = "browser-token"

func (app *App) getBrowserTokenCookie(w http.ResponseWriter,
	r *http.Request) int {
	cookie, err := r.Cookie(COOKIE_NAME)
	if err == nil {
		browserId := db.LookupIdForBrowserToken(app.dbConn, cookie.Value)
		if browserId == 0 {
			log.Printf("No browser row for %s cookie", COOKIE_NAME)
			http.SetCookie(w, &http.Cookie{
				Name:    COOKIE_NAME,
				Expires: time.Unix(0, 0),
			})
		} else {
			db.TouchBrowserLastSeenAt(app.dbConn, browserId)
		}
		return browserId
	} else if err == http.ErrNoCookie {
		return 0
	} else {
		panic(err)
	}
}

func (app *App) getOrSetBrowserTokenCookie(w http.ResponseWriter,
	r *http.Request) int {

	cookie, err := r.Cookie(COOKIE_NAME)
	if err == nil {
		browserId := db.LookupIdForBrowserToken(app.dbConn, cookie.Value)
		if browserId == 0 {
			log.Printf("No browser row for %s cookie", COOKIE_NAME)

			browser := db.InsertIntoBrowsers(app.dbConn, db.BrowsersRow{
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

			return browser.Id
		} else {
			db.TouchBrowserLastSeenAt(app.dbConn, browserId)
		}
		return browserId
	} else if err == http.ErrNoCookie {
		browser := db.InsertIntoBrowsers(app.dbConn, db.BrowsersRow{
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

		return browser.Id
	} else {
		panic(err)
	}
}
