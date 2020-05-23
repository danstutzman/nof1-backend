package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"encoding/json"
	"fmt"
	"gopkg.in/guregu/null.v3"
)

type SyncRequest struct {
	Logs         []map[string]interface{} `json:"logs"`
	LastUpdateId int                      `json:"lastUpdateId"`
	Updates      []db.UpdatesRow          `json:"updates"`
}

type SyncResponse struct {
	Updates []db.UpdatesRow `json:"updates"`
}

func convertClientLogToLogsRow(clientLog map[string]interface{},
	browserId int64) db.LogsRow {

	var idOnClient int
	if f, ok := clientLog["idOnClient"].(float64); ok {
		idOnClient = int(f)
	}
	delete(clientLog, "idOnClient")

	timeOnClient, _ := clientLog["time"].(float64)
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

func (model *Model) PostSync(request SyncRequest,
	userId int64) SyncResponse {

	for _, clientLog := range request.Logs {
		db.InsertIntoLogs(model.dbConn,
			convertClientLogToLogsRow(clientLog, userId))
	}

	for _, update := range request.Updates {
		db.InsertIntoUpdates(model.dbConn, update)
	}

	updates := db.FromUpdates(model.dbConn,
		fmt.Sprintf("WHERE id > %d", request.LastUpdateId))

	return SyncResponse{Updates: updates}
}
