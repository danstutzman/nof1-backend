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
	Path               string
	Prompt             string
}

func assertRecordingsHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, user_id, id_on_client, recorded_at_on_client,
		uploaded_at, path, prompt
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
			(user_id, id_on_client, recorded_at_on_client, uploaded_at, path, prompt)
			VALUES (%d, %d, %f, %d, %s, %s)`,
		row.UserId,
		row.IdOnClient,
		row.RecordedAtOnClient,
		row.UploadedAt.Unix(),
		EscapeString(row.Path),
		EscapeString(row.Prompt))
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}
