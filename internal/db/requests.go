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

func prepareFakeRequests(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE requests (
			id           INTEGER PRIMARY KEY NOT NULL,
			browser_id   INTEGER,
			http_version TEXT NOT NULL,
			tls_protocol TEXT,
			tls_cipher   TEXT,
			received_at  TEXT NOT NULL,
			remote_addr  TEXT NOT NULL,
			method       TEXT NOT NULL,
			path         TEXT NOT NULL,
			duration_ms  INTEGER NOT NULL,
			status_code  INTEGER NOT NULL,
			size         INTEGER NOT NULL,
			error_stack  TEXT
		);`)
	if err != nil {
		log.Fatal(err)
	}
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

func FromRequests(db *sql.DB, whereLimit string) []RequestsRow {
	query := `SELECT id, browser_id, http_version, tls_protocol,
    tls_cipher, received_at, remote_addr, method, path,
		duration_ms, status_code, size, error_stack
    FROM requests ` + whereLimit
	if LOG {
		log.Println(query)
	}

	rset, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rset.Close()

	var rows []RequestsRow
	for rset.Next() {
		var row RequestsRow
		var receivedAt string
		err = rset.Scan(&row.Id,
			&row.BrowserId,
			&row.HttpVersion,
			&row.TlsProtocol,
			&row.TlsCipher,
			&receivedAt,
			&row.RemoteAddr,
			&row.Method,
			&row.Path,
			&row.DurationMs,
			&row.StatusCode,
			&row.Size,
			&row.ErrorStack)
		if err != nil {
			panic(err)
		}

		row.ReceivedAt, err = time.Parse(time.RFC3339, receivedAt)
		if err != nil {
			panic(err)
		}

		rows = append(rows, row)
	}

	err = rset.Err()
	if err != nil {
		panic(err)
	}

	return rows
}
