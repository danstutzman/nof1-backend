package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type RecordingsRow struct {
	Id                 int64
	UserId             int64
	IdOnClient         int
	RecordedAtOnClient float64
	UploadedAt         time.Time
	Filename           string
	MimeType           string
	Size               int
	Prompt             string
}

func assertRecordingsHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, user_id, id_on_client, recorded_at_on_client,
		uploaded_at, filename, mime_type, size, prompt
	  FROM recordings LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoRecordings(db *sql.DB, row RecordingsRow) {
	query := fmt.Sprintf(`INSERT INTO recordings
			(user_id, id_on_client, recorded_at_on_client, uploaded_at, filename,
			mime_type, size, prompt)
			VALUES (%d, %d, %f, %d, %s, %s, %d, %s)`,
		row.UserId,
		row.IdOnClient,
		row.RecordedAtOnClient,
		row.UploadedAt.Unix(),
		EscapeString(row.Filename),
		EscapeString(row.MimeType),
		row.Size,
		EscapeString(row.Prompt))
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func FromRecordings(db *sql.DB, whereLimit string) []RecordingsRow {
	query := `SELECT id, user_id, id_on_client, recorded_at_on_client,
		uploaded_at, filename, mime_type, size, prompt
		FROM recordings ` + whereLimit
	if LOG {
		log.Println(query)
	}

	rset, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rset.Close()

	var rows []RecordingsRow
	for rset.Next() {
		var row RecordingsRow
		var uploadedAt int
		err = rset.Scan(&row.Id,
			&row.UserId,
			&row.IdOnClient,
			&row.RecordedAtOnClient,
			&uploadedAt,
			&row.Filename,
			&row.MimeType,
			&row.Size,
			&row.Prompt)
		if err != nil {
			panic(err)
		}

		row.UploadedAt = time.Unix(int64(uploadedAt), 0)

		rows = append(rows, row)
	}

	err = rset.Err()
	if err != nil {
		panic(err)
	}

	return rows
}
