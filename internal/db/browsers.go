package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type BrowsersRow struct {
	Id             int
	UserAgent      string
	Accept         string
	AcceptEncoding string
	AcceptLanguage string
	Referer        string
	CreatedAt      time.Time
}

func assertBrowsersHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, user_agent, accept, accept_encoding, accept_language,
		referer, created_at
	  FROM browsers LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoBrowsers(db *sql.DB, row BrowsersRow) BrowsersRow {
	row.CreatedAt = time.Now().UTC()

	query := fmt.Sprintf(`INSERT INTO browsers
			(user_agent, accept, accept_encoding, accept_language,
			referer, created_at)
			VALUES (%s, %s, %s, %s,
			 %s, %s)`,
		EscapeString(row.UserAgent),
		EscapeString(row.Accept),
		EscapeString(row.AcceptEncoding),
		EscapeString(row.AcceptLanguage),
		EscapeString(row.Referer),
		EscapeNanoTime(row.CreatedAt))
	if LOG {
		log.Println(query)
	}

	result, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}
	row.Id = int(id)

	return row
}
