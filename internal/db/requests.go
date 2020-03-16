package db

import (
	"database/sql"
	"fmt"
	"gopkg.in/guregu/null.v3"
	"log"
	"time"
)

type RequestsRow struct {
	Id          int64
	ReceivedAt  time.Time
	RemoteAddr  string
	BrowserId   null.Int
	HttpVersion string
	TlsProtocol null.String
	TlsCipher   null.String
	Method      string
	Path        string
	DurationMs  int
	StatusCode  int
	Size        int
	ErrorStack  null.String
}

func assertRequestsHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, received_at, remote_addr, browser_id
	    http_version, tls_protocol, tls_cipher,
  	  method, path,
			duration_ms, status_code, size, error_stack
	  FROM requests LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoRequests(db *sql.DB, row RequestsRow) RequestsRow {
	query := fmt.Sprintf(`INSERT INTO requests
    (received_at, remote_addr, browser_id,
		 http_version, tls_protocol, tls_cipher,
		 method, path,
		 duration_ms, status_code, size, error_stack)
    VALUES (%s, %s, %s,
		 %s, %s, %s,
		 %s, %s,
		 %d, %d, %d, %s)`,
		EscapeNanoTime(row.ReceivedAt),
		EscapeString(row.RemoteAddr),
		EscapeNullInt(row.BrowserId),
		EscapeString(row.HttpVersion),
		EscapeNullString(row.TlsProtocol),
		EscapeNullString(row.TlsCipher),
		EscapeString(row.Method),
		EscapeString(row.Path),
		row.DurationMs,
		row.StatusCode,
		row.Size,
		EscapeNullString(row.ErrorStack))
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
	row.Id = id

	return row
}
