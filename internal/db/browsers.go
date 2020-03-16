package db

import (
	"database/sql"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"gopkg.in/guregu/null.v3"
	"log"
	"time"
)

type BrowsersRow struct {
	Id             int64
	Token          string
	UserAgent      string
	Accept         string
	AcceptEncoding string
	AcceptLanguage string
	Referer        string
	UserId         null.Int
	CreatedAt      time.Time
	LastSeenAt     time.Time
}

func assertBrowsersHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, token, user_agent, accept, accept_encoding,
		accept_language, referer, user_id, created_at, last_seen_at
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
				referer, user_id, created_at, last_seen_at)
				VALUES (%s, %s, %s, %s, %s,
				 %s, %s, %d, %d)`,
			EscapeString(row.Token),
			EscapeString(row.UserAgent),
			EscapeString(row.Accept),
			EscapeString(row.AcceptEncoding),
			EscapeString(row.AcceptLanguage),
			EscapeString(row.Referer),
			EscapeNullInt(row.UserId),
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
			row.Id = id
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

func FromBrowsers(db *sql.DB, whereLimit string) []BrowsersRow {
	query := `SELECT id, token, user_agent, accept, accept_encoding,
		accept_language, referer, user_id, created_at, last_seen_at
	  FROM browsers ` + whereLimit
	if LOG {
		log.Println(query)
	}

	rset, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rset.Close()

	var rows []BrowsersRow
	for rset.Next() {
		var row BrowsersRow
		var createdAt int
		var lastSeenAt int
		err = rset.Scan(&row.Id,
			&row.Token,
			&row.UserAgent,
			&row.Accept,
			&row.AcceptEncoding,
			&row.AcceptLanguage,
			&row.Referer,
			&row.UserId,
			&createdAt,
			&lastSeenAt)
		if err != nil {
			panic(err)
		}

		row.CreatedAt = time.Unix(int64(createdAt), 0)
		row.LastSeenAt = time.Unix(int64(lastSeenAt), 0)

		rows = append(rows, row)
	}

	err = rset.Err()
	if err != nil {
		panic(err)
	}

	return rows
}

func TouchBrowserLastSeenAt(db *sql.DB, browserId int64) {
	query := "UPDATE browsers SET last_seen_at = $1 WHERE id = $2"
	if LOG {
		log.Println(query)
	}

	lastSeenAt := time.Now().UTC()
	_, err := db.Exec(query, lastSeenAt.Unix(), browserId)
	if err != nil {
		panic(err)
	}
}

func UpdateUserIdAndLastSeenAtOnBrowser(db *sql.DB, userId, browserId int64) {
	query := "UPDATE browsers SET user_id = $1, last_seen_at = $2 WHERE id = $3"
	if LOG {
		log.Println(query)
	}

	lastSeenAt := time.Now().UTC()
	_, err := db.Exec(query, userId, lastSeenAt.Unix(), browserId)
	if err != nil {
		panic(err)
	}
}
