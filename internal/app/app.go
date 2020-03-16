package app

import (
	"database/sql"
)

type App struct {
	dbConn    *sql.DB
	uploadDir string
}

func NewApp(
	dbConn *sql.DB,
	uploadDir string,
) *App {
	return &App{
		dbConn:    dbConn,
		uploadDir: uploadDir,
	}
}
