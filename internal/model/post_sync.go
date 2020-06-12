package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"gopkg.in/guregu/null.v3"
)

type SyncRequest struct {
	SyncedUntilDeltaId int64                    `json:"syncedUntilDeltaId"`
	Deltas             []map[string]interface{} `json:"deltas"`
}

type SyncResponse struct {
	SyncedUntilDeltaId int64          `json:"syncedUntilDeltaId"`
	Deltas             []db.DeltasRow `json:"deltas"`
}

func convertDeltaToLogsRow(delta map[string]interface{},
	browserId int64) db.LogsRow {

	var idOnClient int
	if f, ok := delta["idOnClient"].(float64); ok {
		idOnClient = int(f)
	}
	delete(delta, "idOnClient")

	timeOnClient, _ := delta["time"].(float64)
	delete(delta, "time")

	message := delta["message"].(string)
	delete(delta, "message")

	var errorName null.String
	var errorMessage null.String
	var errorStack null.String
	if delta["error"] != nil {
		if errorMap, ok := delta["error"].(map[string]interface{}); ok {
			if s, ok := errorMap["name"].(string); ok {
				errorName = null.StringFrom(s)
			}
			if s, ok := errorMap["message"].(string); ok {
				errorMessage = null.StringFrom(s)
			}
			if s, ok := errorMap["stack"].(string); ok {
				errorStack = null.StringFrom(s)
			}
			delete(delta, "error")
			delete(delta, "error")
		}
	}

	var otherDetailsJson null.String
	if len(delta) > 0 {
		json, err := json.Marshal(delta)
		if err != nil {
			panic(err)
		}
		otherDetailsJson = null.StringFrom(string(json))
	}

	return db.LogsRow{
		BrowserId:        browserId,
		IdOnClient:       idOnClient,
		TimeOnClient:     timeOnClient,
		Message:          message,
		ErrorName:        errorName,
		ErrorMessage:     errorMessage,
		ErrorStack:       errorStack,
		OtherDetailsJson: otherDetailsJson,
	}
}

func handleDeltaTypeLog(dbConn *sql.DB, delta map[string]interface{},
	browserId int64) {

	db.InsertIntoLogs(dbConn, convertDeltaToLogsRow(delta, browserId))
}

func handleDeltaTypeUpdateRecordingTranscriptManual(dbConn *sql.DB,
	delta map[string]interface{}) (int64, error) {

	idOnClient, ok := delta["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("Couldn't convert id to float64")
	}

	timeOnClient, ok := delta["time"].(float64)
	if !ok {
		return 0, fmt.Errorf("Couldn't convert timeOnClient to float64")
	}

	transcriptManual, ok := delta["transcriptManual"].(string)
	if !ok {
		return 0, fmt.Errorf("Couldn't convert transcriptManual to string")
	}

	recordingIdOnServer, ok := delta["recordingIdOnServer"].(float64)
	if !ok {
		return 0, fmt.Errorf("Couldn't convert recordingIdOnServer to float64")
	}

	db.UpdateTranscriptManualOnRecording(
		dbConn, transcriptManual, int64(recordingIdOnServer))

	newDelta := db.InsertIntoDeltas(dbConn, db.DeltasRow{
		Type:             db.DELTA_TYPE_UPDATE_RECORDING_TRANSCRIPT_MANUAL,
		IdOnClient:       null.IntFrom(int64(idOnClient)),
		TimeOnClient:     null.FloatFrom(timeOnClient),
		RecordingId:      null.IntFrom(int64(recordingIdOnServer)),
		TranscriptManual: null.StringFrom(transcriptManual),
	})

	return newDelta.Id, nil
}

func (model *Model) PostSync(request SyncRequest,
	browserId int64) (*SyncResponse, error) {

	newSyncedUntilDeltaId := request.SyncedUntilDeltaId

	deltas := db.FromDeltas(model.dbConn,
		fmt.Sprintf("WHERE id > %d", request.SyncedUntilDeltaId))

	for _, delta := range request.Deltas {
		switch delta["type"] {
		case db.DELTA_TYPE_LOG:
			handleDeltaTypeLog(model.dbConn, delta, browserId)
		case db.DELTA_TYPE_UPDATE_RECORDING_TRANSCRIPT_MANUAL:
			newDeltaId, err :=
				handleDeltaTypeUpdateRecordingTranscriptManual(model.dbConn, delta)
			if err != nil {
				return nil, err
			}
			newSyncedUntilDeltaId = newDeltaId
		default:
			return nil, fmt.Errorf("Unexpected delta type on %+v", delta)
		}
	}

	return &SyncResponse{
		SyncedUntilDeltaId: newSyncedUntilDeltaId,
		Deltas:             deltas,
	}, nil
}
