package config

import "fmt"

// Config holds all application configuration.
type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	DBSchema           string
	DBTable            string
	PubSubSubscription string
	CreateTable        bool
	HealthPort         string
}

// PostgresConnStr returns a PostgreSQL connection string.
func (c *Config) PostgresConnStr() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// Validate checks that required fields are set.
func (c *Config) Validate() error {
	if c.DBTable == "" {
		return fmt.Errorf("db-table is required")
	}
	if c.PubSubSubscription == "" {
		return fmt.Errorf("pubsub-subscription is required")
	}
	return nil
}
