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
	MetadataJson       string
	TranscriptAws      string
	TranscriptManual   string
	AwsTranscribeJson  string
}

func assertRecordingsHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, user_id, id_on_client, recorded_at_on_client,
		uploaded_at, filename, mime_type, size, metadata_json, transcript_aws,
		transcript_manual, aws_transcribe_json
	  FROM recordings LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoRecordings(db *sql.DB, row RecordingsRow) RecordingsRow {
	query := fmt.Sprintf(`INSERT INTO recordings
			(user_id, id_on_client, recorded_at_on_client, uploaded_at, filename,
			mime_type, size, metadata_json, transcript_aws, transcript_manual,
			aws_transcribe_json)
			VALUES (%d, %d, %f, %d, %s, %s, %d, %s, %s, %s, %s)`,
		row.UserId,
		row.IdOnClient,
		row.RecordedAtOnClient,
		row.UploadedAt.Unix(),
		EscapeString(row.Filename),
		EscapeString(row.MimeType),
		row.Size,
		EscapeString(row.MetadataJson),
		EscapeString(row.TranscriptAws),
		EscapeString(row.TranscriptManual),
		EscapeString(row.AwsTranscribeJson))
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

func FromRecordings(db *sql.DB, whereLimit string) []RecordingsRow {
	query := `SELECT id, user_id, id_on_client, recorded_at_on_client,
		uploaded_at, filename, mime_type, size, metadata_json, transcript_aws,
		transcript_manual, aws_transcribe_json
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
			&row.MetadataJson,
			&row.TranscriptAws,
			&row.TranscriptManual,
			&row.AwsTranscribeJson)
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

func UpdateTranscriptManualOnRecording(db *sql.DB, transcriptManual string,
	recordingId int64) {

	query := "UPDATE recordings SET transcript_manual = $1 WHERE id = $2"
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query, transcriptManual, recordingId)
	if err != nil {
		panic(err)
	}
}

func UpdateTranscriptAwsOnRecording(db *sql.DB, transcriptAws string,
	awsTranscribeJson string, recordingId int64) {

	query := `UPDATE recordings
		SET transcript_aws = $1, aws_transcribe_json = $2
		WHERE id = $3`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query, transcriptAws, awsTranscribeJson, recordingId)
	if err != nil {
		panic(err)
	}
}
