package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"time"
)

type BrowsersRow struct {
	Id             int
	Token          string
	UserAgent      string
	Accept         string
	AcceptEncoding string
	AcceptLanguage string
	Referer        string
	CreatedAt      time.Time
	LastSeenAt     time.Time
}

func generateToken() string {
	buffer := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, buffer)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}

func assertBrowsersHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, token, user_agent, accept, accept_encoding,
		accept_language, referer, created_at
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
	for numCollisions := 0; numCollisions < 10; numCollisions += 1 {
		row.Token = generateToken()
		row.CreatedAt = time.Now().UTC()
		row.LastSeenAt = time.Now().UTC()

		query := fmt.Sprintf(`INSERT INTO browsers
				(token, user_agent, accept, accept_encoding, accept_language,
				referer, created_at, last_seen_at)
				VALUES (%s, %s, %s, %s, %s,
				 %s, %d, %d)`,
			EscapeString(row.Token),
			EscapeString(row.UserAgent),
			EscapeString(row.Accept),
			EscapeString(row.AcceptEncoding),
			EscapeString(row.AcceptLanguage),
			EscapeString(row.Referer),
			row.CreatedAt.Unix(),
			row.LastSeenAt.Unix())
		if LOG {
			log.Println(query)
		}

		result, err := db.Exec(query)
		if err == nil {
			id, err := result.LastInsertId()
			if err != nil {
				panic(err)
			}
			row.Id = int(id)
			return row
		} else if sqliteErr, ok := err.(sqlite3.Error); ok &&
			sqliteErr.Code == sqlite3.ErrConstraint {
			// Let the loop repeat
		} else {
			panic(err)
		}
	} // Loop
	panic("Too many collisions")
}

func LookupIdForBrowserToken(db *sql.DB, token string) int {
	query := "SELECT id FROM browsers WHERE token = $1"
	if LOG {
		log.Println(query)
	}

	var id int
	row := db.QueryRow(query, token)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return 0
	} else if err == nil {
		return id
	} else {
		panic(err)
	}
}

func TouchBrowserLastSeenAt(db *sql.DB, browserId int) {
	query := "UPDATE browsers SET last_seen_at = $1 WHERE id = $2"
	if LOG {
		log.Println(query)
	}

	lastSeenAt := time.Now().UTC()
	_, err := db.Exec(query, EscapeNanoTime(lastSeenAt), browserId)
	if err != nil {
		panic(err)
	}
}
