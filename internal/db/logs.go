package db

import (
	"database/sql"
	"fmt"
	"gopkg.in/guregu/null.v3"
	"log"
)

type LogsRow struct {
	BrowserId        int64
	IdOnClient       int
	TimeOnClient     float64
	Message          string
	ErrorName        null.String
	ErrorMessage     null.String
	ErrorStack       null.String
	OtherDetailsJson null.String
}

func assertLogsHasCorrectSchema(db *sql.DB) {
	query := `SELECT browser_id, id_on_client, time_on_client, message,
		error_name, error_message, error_stack, other_details_json
	  FROM logs LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoLogs(db *sql.DB, row LogsRow) {
	query := fmt.Sprintf(`INSERT INTO logs
    (browser_id, id_on_client, time_on_client, message,
		 error_name, error_message, error_stack, other_details_json)
    VALUES (%d, %d, %f, %s,
		 %s, %s, %s, %s)`,
		row.BrowserId,
		row.IdOnClient,
		row.TimeOnClient,
		EscapeString(row.Message),
		EscapeNullString(row.ErrorName),
		EscapeNullString(row.ErrorMessage),
		EscapeNullString(row.ErrorStack),
		EscapeNullString(row.OtherDetailsJson))
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}
