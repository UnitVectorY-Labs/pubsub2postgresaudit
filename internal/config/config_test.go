package config

import (
	"testing"
)

func TestValidate_MissingDBTable(t *testing.T) {
	cfg := &Config{
		PubSubSubscription: "projects/p/subscriptions/s",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing db-table")
	}
}

func TestValidate_MissingPubSubSubscription(t *testing.T) {
	cfg := &Config{
		DBTable: "audit",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing pubsub-subscription")
	}
}

func TestValidate_AllSet(t *testing.T) {
	cfg := &Config{
		DBTable:            "audit",
		PubSubSubscription: "projects/p/subscriptions/s",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresConnStr(t *testing.T) {
	cfg := &Config{
		DBHost:     "myhost",
		DBPort:     "5433",
		DBUser:     "admin",
		DBPassword: "secret",
		DBName:     "mydb",
		DBSSLMode:  "require",
	}
	want := "host=myhost port=5433 user=admin password=secret dbname=mydb sslmode=require"
	got := cfg.PostgresConnStr()
	if got != want {
		t.Fatalf("PostgresConnStr()\n got: %s\nwant: %s", got, want)
	}
}
