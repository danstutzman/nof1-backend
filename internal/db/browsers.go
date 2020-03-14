package db

import (
	"database/sql"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
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

func init() {
	rand.Seed(time.Now().UnixNano())
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
	inserted := false
	for !inserted {
		row.Id = int(rand.Int31())
		for row.Id == 0 {
			row.Id = int(rand.Int31())
		}

		row.CreatedAt = time.Now().UTC()

		query := fmt.Sprintf(`INSERT INTO browsers
			(id, user_agent, accept, accept_encoding, accept_language,
			referer, created_at)
			VALUES (%d, %s, %s, %s, %s,
			 %s, %s)`,
			row.Id,
			EscapeString(row.UserAgent),
			EscapeString(row.Accept),
			EscapeString(row.AcceptEncoding),
			EscapeString(row.AcceptLanguage),
			EscapeString(row.Referer),
			EscapeNanoTime(row.CreatedAt))
		if LOG {
			log.Println(query)
		}

		_, err := db.Exec(query)
		if err == nil {
			inserted = true
		} else if sqliteErr, ok := err.(sqlite3.Error); ok &&
			sqliteErr.Code == sqlite3.ErrConstraint {
			inserted = false
		} else {
			panic(err)
		}
	}

	return row
}
