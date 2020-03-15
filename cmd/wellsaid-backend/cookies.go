package main

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (app *App) getBrowserIdCookie(w http.ResponseWriter, r *http.Request) int {
	cookie, err := r.Cookie("browser-id")
	if err == nil {
		decrypted, err := decrypt(cookie.Value, app.secretKey)
		if err != nil {
			log.Printf("Couldn't decrypt cookie: %v", err)
			http.SetCookie(w, &http.Cookie{
				Name:    "browser-id",
				Expires: time.Unix(0, 0),
			})
			return 0
		}
		browserId, _ := strconv.Atoi(decrypted)
		return browserId
	} else if err == http.ErrNoCookie {
		return 0
	} else {
		panic(err)
	}
}

func (app *App) getOrSetBrowserIdCookie(w http.ResponseWriter,
	r *http.Request) int {

	cookie, err := r.Cookie("browser-id")
	if err == nil {
		decrypted, err := decrypt(cookie.Value, app.secretKey)
		if err != nil {
			log.Printf("Couldn't decrypt cookie: %v", err)
			http.SetCookie(w, &http.Cookie{
				Name:    "browser-id",
				Expires: time.Unix(0, 0),
			})
			return 0
		}
		browserId, _ := strconv.Atoi(decrypted)
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
			Name:    "browser-id",
			Value:   encrypt(strconv.Itoa(browser.Id), app.secretKey),
			Expires: time.Now().AddDate(30, 0, 0),
		})

		return browser.Id
	} else {
		panic(err)
	}
}
