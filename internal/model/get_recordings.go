package model

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
)

func (model *Model) GetRecordings() []db.RecordingsRow {
	return db.FromRecordings(model.dbConn, "")
}
