package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

const LOG = false

func InitDb(dbPath string) *sql.DB {
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		log.Fatalf("File doesn't exist: %s", dbPath)
	} else if err != nil {
		log.Fatalf("Error with file %s", err)
	}

	// Set mode=rw so it doesn't create database if file doesn't exist
	connString := "file:" + dbPath + "?mode=rw"
	dbConn, err := sql.Open("sqlite3", connString)
	if err != nil {
		log.Fatalf("Error from sql.Open: %s", err)
	}

	assertRequestsHasCorrectSchema(dbConn)

	return dbConn
}
