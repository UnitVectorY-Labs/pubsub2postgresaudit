package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/config"
	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/database"
)

// ParseSubscription extracts project ID and subscription ID from
// a full subscription name of the form "projects/<project>/subscriptions/<sub>".
func ParseSubscription(full string) (projectID, subscriptionID string, err error) {
	parts := strings.Split(full, "/")
	if len(parts) != 4 || parts[0] != "projects" || parts[2] != "subscriptions" || parts[1] == "" || parts[3] == "" {
		return "", "", fmt.Errorf("invalid subscription name: %q", full)
	}
	return parts[1], parts[3], nil
}

// Run consumes Pub/Sub messages and inserts them into PostgreSQL.
func Run(ctx context.Context, cfg *config.Config, db *sql.DB) error {
	projectID, subID, err := ParseSubscription(cfg.PubSubSubscription)
	if err != nil {
		return err
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("creating pubsub client: %w", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	slog.Info("starting consumer", "subscription", cfg.PubSubSubscription)

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		handleMessage(ctx, cfg, db, msg)
	})
	if err != nil {
		slog.Error("pubsub receive error",
			"subscription", cfg.PubSubSubscription,
			"outcome", "pubsub_error",
			"error", err.Error(),
		)
		return fmt.Errorf("receiving messages: %w", err)
	}
	return nil
}

func handleMessage(ctx context.Context, cfg *config.Config, db *sql.DB, msg *pubsub.Message) {
	logAttrs := []any{
		"subscription", cfg.PubSubSubscription,
		"message_id", msg.ID,
		"publish_time", msg.PublishTime.Format(time.RFC3339),
		"db_schema", cfg.DBSchema,
		"db_table", cfg.DBTable,
	}

	// Marshal attributes as JSON
	attrJSON, err := json.Marshal(msg.Attributes)
	if err != nil {
		slog.Error("failed to marshal attributes",
			append(logAttrs, "outcome", "invalid_json", "error", err.Error(), "byte_size", len(msg.Data))...,
		)
		msg.Ack()
		return
	}

	// Validate data is valid JSON
	if !json.Valid(msg.Data) {
		slog.Warn("invalid JSON data",
			append(logAttrs, "outcome", "invalid_json", "error", "data is not valid JSON", "byte_size", len(msg.Data))...,
		)
		msg.Ack()
		return
	}

	inserted, err := database.InsertMessage(ctx, db, cfg.DBSchema, cfg.DBTable, msg.ID, msg.PublishTime, json.RawMessage(attrJSON), json.RawMessage(msg.Data))
	if err != nil {
		slog.Error("database insert error",
			append(logAttrs, "outcome", "db_error", "error", err.Error())...,
		)
		msg.Nack()
		return
	}

	if inserted {
		slog.Info("message inserted", append(logAttrs, "outcome", "inserted")...)
	} else {
		slog.Info("duplicate message", append(logAttrs, "outcome", "duplicate")...)
	}
	msg.Ack()
}
