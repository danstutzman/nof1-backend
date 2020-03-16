package db

import (
	"database/sql"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type UsersRow struct {
	Id         int64
	Token      string
	CreatedAt  time.Time
	LastSeenAt time.Time
}

func assertUsersHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, token, created_at, last_seen_at
	  FROM users LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

// Returns ID of newly created user
func InsertIntoUsers(db *sql.DB) int64 {
	for numCollisions := 0; numCollisions < 10; numCollisions += 1 {
		token := generateToken()
		createdAt := time.Now().UTC()
		lastSeenAt := time.Now().UTC()

		query := fmt.Sprintf(`INSERT INTO users
				(token, created_at, last_seen_at)
				VALUES (%s, %d, %d)`,
			EscapeString(token),
			createdAt.Unix(),
			lastSeenAt.Unix())
		if LOG {
			log.Println(query)
		}

		result, err := db.Exec(query)
		if err == nil {
			id, err := result.LastInsertId()
			if err != nil {
				panic(err)
			}
			return id
		} else if sqliteErr, ok := err.(sqlite3.Error); ok &&
			sqliteErr.Code == sqlite3.ErrConstraint {
			// Let the loop repeat
		} else {
			panic(err)
		}
	} // Loop
	panic("Too many collisions")
}

func LookupIdForUserToken(db *sql.DB, token string) int {
	query := "SELECT id FROM users WHERE token = $1"
	if LOG {
		log.Println(query)
	}

	var id int
	row := db.QueryRow(query, token)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return 0
	} else if err == nil {
		return id
	} else {
		panic(err)
	}
}

func TouchUserLastSeenAt(db *sql.DB, userId int) {
	query := "UPDATE users SET last_seen_at = $1 WHERE id = $2"
	if LOG {
		log.Println(query)
	}

	lastSeenAt := time.Now().UTC()
	_, err := db.Exec(query, lastSeenAt.Unix(), userId)
	if err != nil {
		panic(err)
	}
}
