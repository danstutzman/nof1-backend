package db

import (
	"database/sql"
	"fmt"
	"gopkg.in/guregu/null.v3"
	"log"
)

const DELTA_TYPE_LOG = "Log"
const DELTA_TYPE_UPDATE_RECORDING_TRANSCRIPT_MANUAL = "UpdateRecordingTranscriptManual"

type DeltasRow struct {
	Id               int64       `json:"id"`
	Type             string      `json:"type"`
	IdOnClient       null.Int    `json:"idOnClient"`
	TimeOnClient     null.Float  `json:"timeOnClient"`
	RecordingId      null.Int    `json:"recordingId"`
	TranscriptManual null.String `json:"transcriptManual"`
	TranscriptAws    null.String `json:"transcriptAws"`
}

func assertDeltasHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, type, id_on_client, time_on_client, recording_id,
	  transcript_manual, transcript_aws
		FROM deltas
		LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoDeltas(db *sql.DB, row DeltasRow) DeltasRow {
	query := fmt.Sprintf(`INSERT INTO deltas
			(type, id_on_client, time_on_client, recording_id, transcript_manual,
		  transcript_aws)
			VALUES (%s, %s, %s, %s, %s, %s)`,
		EscapeString(row.Type),
		EscapeNullInt(row.IdOnClient),
		EscapeNullFloat(row.TimeOnClient),
		EscapeNullInt(row.RecordingId),
		EscapeNullString(row.TranscriptManual),
		EscapeNullString(row.TranscriptAws))
	if LOG {
		log.Println(query)
	}

	result, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	row.Id, err = result.LastInsertId()
	if err != nil {
		panic(err)
	}

	return row
}

func FromDeltas(db *sql.DB, whereLimit string) []DeltasRow {
	query := `SELECT id, type, id_on_client, time_on_client, recording_id,
		transcript_manual, transcript_aws
    FROM deltas ` + whereLimit
	if LOG {
		log.Println(query)
	}

	rset, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rset.Close()

	rows := []DeltasRow{}
	for rset.Next() {
		var row DeltasRow
		err = rset.Scan(&row.Id,
			&row.Type,
			&row.IdOnClient,
			&row.TimeOnClient,
			&row.RecordingId,
			&row.TranscriptManual,
			&row.TranscriptAws)
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
