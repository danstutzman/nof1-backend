package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type UpdatesRow struct {
	Id         int64     `json:"id"`
	TableName  string    `json:"tableName"`
	RowId      int64     `json:"rowId"`
	ColumnName string    `json:"columnName"`
	NewValue   string    `json:"newValue"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func assertUpdatesHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, table_name, row_id, row_new_value, updated_at
		FROM updates
		LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoUpdates(db *sql.DB, row UpdatesRow) UpdatesRow {
	query := fmt.Sprintf(`INSERT INTO updates
			(table_name, row_id, column_name, new_value, updated_at)
			VALUES (%s, %d, %s, %s, %d)`,
		EscapeString(row.TableName),
		row.RowId,
		EscapeString(row.ColumnName),
		EscapeString(row.NewValue),
		row.UpdatedAt.Unix())
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

func FromUpdates(db *sql.DB, whereLimit string) []UpdatesRow {
	query := `SELECT id, table_name, row_id, column_name, new_value, updated_at
    FROM updates ` + whereLimit
	if LOG {
		log.Println(query)
	}

	rset, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rset.Close()

	rows := []UpdatesRow{}
	for rset.Next() {
		var row UpdatesRow
		var updatedAt int
		err = rset.Scan(&row.Id,
			&row.TableName,
			&row.RowId,
			&row.ColumnName,
			&row.NewValue,
			&updatedAt)
		if err != nil {
			panic(err)
		}

		row.UpdatedAt = time.Unix(int64(updatedAt), 0)

		rows = append(rows, row)
	}

	err = rset.Err()
	if err != nil {
		panic(err)
	}

	return rows
}
