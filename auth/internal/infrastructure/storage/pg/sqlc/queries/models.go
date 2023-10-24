// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package queries

import (
	"database/sql"
)

type ApiKey struct {
	ID          string
	IsRoot      sql.NullBool
	RateLimitID sql.NullInt32
	ServiceName sql.NullString
}

type RateLimit struct {
	ID               int32
	RequestsPerRange sql.NullInt32
	RangeInSeconds   sql.NullInt32
}
