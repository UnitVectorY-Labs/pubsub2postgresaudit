package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	_ "github.com/lib/pq"
)

var validIdentifier = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// ValidateIdentifier checks that s contains only alphanumeric characters and underscores.
func ValidateIdentifier(s string) error {
	if !validIdentifier.MatchString(s) {
		return fmt.Errorf("invalid identifier: %q", s)
	}
	return nil
}

// quoteIdentifier returns the identifier wrapped in double quotes.
func quoteIdentifier(s string) string {
	return `"` + s + `"`
}

// Connect opens and pings a PostgreSQL connection.
func Connect(ctx context.Context, connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return db, nil
}

// Migrate creates the audit table if it does not exist.
func Migrate(ctx context.Context, db *sql.DB, schema, table string) error {
	if err := ValidateIdentifier(schema); err != nil {
		return err
	}
	if err := ValidateIdentifier(table); err != nil {
		return err
	}

	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s.%s (
			message_id TEXT PRIMARY KEY,
			publish_time TIMESTAMPTZ NOT NULL,
			attributes JSONB NOT NULL,
			data JSONB NOT NULL
		)`,
		quoteIdentifier(schema),
		quoteIdentifier(table),
	)

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}
	return nil
}

// InsertMessage inserts a Pub/Sub message into the audit table.
// On conflict (duplicate message_id) it returns inserted=false, err=nil.
func InsertMessage(ctx context.Context, db *sql.DB, schema, table, messageID string, publishTime time.Time, attributes, data json.RawMessage) (bool, error) {
	if err := ValidateIdentifier(schema); err != nil {
		return false, err
	}
	if err := ValidateIdentifier(table); err != nil {
		return false, err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s.%s (message_id, publish_time, attributes, data)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (message_id) DO NOTHING`,
		quoteIdentifier(schema),
		quoteIdentifier(table),
	)

	result, err := db.ExecContext(ctx, query, messageID, publishTime, attributes, data)
	if err != nil {
		return false, fmt.Errorf("inserting message: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("checking rows affected: %w", err)
	}
	return rows > 0, nil
}
