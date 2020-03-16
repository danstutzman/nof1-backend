package app

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"encoding/json"
	"gopkg.in/guregu/null.v3"
)

type SyncRequest struct {
	Logs []map[string]interface{}
}

func convertClientLogToLogsRow(clientLog map[string]interface{},
	userId int) db.LogsRow {

	var idOnClient int
	if f, ok := clientLog["id"].(float64); ok {
		idOnClient = int(f)
	}
	delete(clientLog, "id")

	var timeOnClient int
	if f, ok := clientLog["time"].(float64); ok {
		timeOnClient = int(f)
	}
	delete(clientLog, "time")

	message := clientLog["message"].(string)
	delete(clientLog, "message")

	var errorName null.String
	var errorMessage null.String
	var errorStack null.String
	if clientLog["error"] != nil {
		if errorMap, ok := clientLog["error"].(map[string]interface{}); ok {
			if s, ok := errorMap["name"].(string); ok {
				errorName = null.StringFrom(s)
			}
			if s, ok := errorMap["message"].(string); ok {
				errorMessage = null.StringFrom(s)
			}
			if s, ok := errorMap["stack"].(string); ok {
				errorStack = null.StringFrom(s)
			}
			delete(clientLog, "error")
			delete(clientLog, "error")
		}
	}

	var otherDetailsJson null.String
	if len(clientLog) > 0 {
		json, err := json.Marshal(clientLog)
		if err != nil {
			panic(err)
		}
		otherDetailsJson = null.StringFrom(string(json))
	}

	return db.LogsRow{
		UserId:           userId,
		IdOnClient:       idOnClient,
		TimeOnClient:     timeOnClient,
		Message:          message,
		ErrorName:        errorName,
		ErrorMessage:     errorMessage,
		ErrorStack:       errorStack,
		OtherDetailsJson: otherDetailsJson,
	}
}

func (app *App) PostSync(syncRequest SyncRequest) {
	userId := 0 // TODO: convert from user token

	for _, clientLog := range syncRequest.Logs {
		db.InsertIntoLogs(app.dbConn,
			convertClientLogToLogsRow(clientLog, userId))
	}
}
