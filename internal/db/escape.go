package db

import (
	"gopkg.in/guregu/null.v3"
	"strconv"
	"strings"
	"time"
)

func EscapeString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func EscapePtr(s *string) string {
	if s == nil {
		return "NULL"
	}

	return EscapeString(*s)
}

func EscapeNullString(s null.String) string {
	if !s.Valid {
		return "NULL"
	}

	return EscapeString(s.String)
}

func EscapeNanoTime(nanoTime time.Time) string {
	return "'" + nanoTime.UTC().Format(time.RFC3339Nano) + "'"
}

func EscapeBool(b bool) string {
	if b {
		return "TRUE"
	} else {
		return "FALSE"
	}
}

func EscapeNullBool(b null.Bool) string {
	if !b.Valid {
		return "NULL"
	}

	return EscapeBool(b.Bool)
}

func EscapeNullInt(i null.Int) string {
	if !i.Valid {
		return "NULL"
	}

	return strconv.FormatInt(i.Int64, 10)
}
