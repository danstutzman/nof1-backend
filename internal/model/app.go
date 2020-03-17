package model

import (
	"database/sql"
)

type Model struct {
	dbConn    *sql.DB
	uploadDir string
}

func NewModel(
	dbConn *sql.DB,
	uploadDir string,
) *Model {
	return &Model{
		dbConn:    dbConn,
		uploadDir: uploadDir,
	}
}
