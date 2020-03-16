package app

import (
	"bitbucket.org/danstutzman/wellsaid-backend/internal/db"
	"encoding/json"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"net/http"
	"time"
)

type SyncRequest struct {
	Logs []map[string]interface{}
}

func convertClientLogToLogsRow(clientLog map[string]interface{},
	browserId int) db.LogsRow {

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

func (app *App) postSync(w http.ResponseWriter, r *http.Request) {
	receivedAt := time.Now().UTC()
	browserId := app.getBrowserTokenCookie(r)
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=\"utf-8\"")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var syncRequest SyncRequest
	err = json.Unmarshal(body, &syncRequest)
	if err != nil {
		panic(err)
	}

	for _, clientLog := range syncRequest.Logs {
		db.InsertIntoLogs(app.dbConn,
			convertClientLogToLogsRow(clientLog, browserId))
	}

	if browserId == 0 {
		browserId = app.setBrowserTokenCookie(w, r)
	}
	bytes := "{}"
	w.Write([]byte(bytes))

	app.logRequest(receivedAt, r, http.StatusOK, len(bytes), null.String{},
		browserId)
}
