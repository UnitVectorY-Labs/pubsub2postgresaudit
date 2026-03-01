package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/config"
	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/consumer"
	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/database"
	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/health"
)

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	// Set up structured JSON logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Check for version flag before anything else
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(Version)
		os.Exit(0)
	}

	// Determine subcommand
	subcommand := ""
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "migrate" {
		subcommand = "migrate"
		args = args[1:]
	}

	fs := flag.NewFlagSet("pubsub2postgresaudit", flag.ExitOnError)

	dbHost := flagOrEnv(fs, "db-host", "DB_HOST", "localhost", "PostgreSQL hostname")
	dbPort := flagOrEnv(fs, "db-port", "DB_PORT", "5432", "PostgreSQL port")
	dbUser := flagOrEnv(fs, "db-user", "DB_USER", "postgres", "Database user")
	dbPassword := flagOrEnv(fs, "db-password", "DB_PASSWORD", "", "Database password")
	dbName := flagOrEnv(fs, "db-name", "DB_NAME", "cert_observatory", "Database name")
	dbSSLMode := flagOrEnv(fs, "db-sslmode", "DB_SSLMODE", "disable", "SSL mode")
	dbSchema := flagOrEnv(fs, "db-schema", "DB_SCHEMA", "public", "Database schema")
	dbTable := flagOrEnv(fs, "db-table", "DB_TABLE", "", "Table name (required)")
	pubsubSub := flagOrEnv(fs, "pubsub-subscription", "PUBSUB_SUBSCRIPTION", "", "Full Pub/Sub subscription name (required)")
	createTable := fs.Bool("create-table", false, "Create table if missing")
	healthPort := flagOrEnv(fs, "health-port", "HEALTH_PORT", "", "Port for health endpoints")

	fs.Parse(args)

	// Apply env var for create-table boolean
	if !isFlagSet(fs, "create-table") {
		if v := os.Getenv("CREATE_TABLE"); v == "true" || v == "1" {
			*createTable = true
		}
	}

	cfg := &config.Config{
		DBHost:             *dbHost,
		DBPort:             *dbPort,
		DBUser:             *dbUser,
		DBPassword:         *dbPassword,
		DBName:             *dbName,
		DBSSLMode:          *dbSSLMode,
		DBSchema:           *dbSchema,
		DBTable:            *dbTable,
		PubSubSubscription: *pubsubSub,
		CreateTable:        *createTable,
		HealthPort:         *healthPort,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if subcommand == "migrate" {
		if err := cfg.Validate(); err != nil {
			slog.Error("configuration error", "error", err.Error())
			os.Exit(1)
		}
		db, err := database.Connect(ctx, cfg.PostgresConnStr())
		if err != nil {
			slog.Error("database connection error", "error", err.Error())
			os.Exit(1)
		}
		defer db.Close()

		if err := database.Migrate(ctx, db, cfg.DBSchema, cfg.DBTable); err != nil {
			slog.Error("migration error", "error", err.Error())
			os.Exit(1)
		}
		slog.Info("migration complete", "schema", cfg.DBSchema, "table", cfg.DBTable)
		return
	}

	// Default: run consumer
	if err := cfg.Validate(); err != nil {
		slog.Error("configuration error", "error", err.Error())
		os.Exit(1)
	}

	db, err := database.Connect(ctx, cfg.PostgresConnStr())
	if err != nil {
		slog.Error("database connection error", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	if cfg.CreateTable {
		if err := database.Migrate(ctx, db, cfg.DBSchema, cfg.DBTable); err != nil {
			slog.Error("table creation error", "error", err.Error())
			os.Exit(1)
		}
		slog.Info("table ensured", "schema", cfg.DBSchema, "table", cfg.DBTable)
	}

	if cfg.HealthPort != "" {
		checker := &health.Checker{DB: db}
		if err := checker.Start(cfg.HealthPort); err != nil {
			slog.Error("health server error", "error", err.Error())
			os.Exit(1)
		}
		slog.Info("health server started", "port", cfg.HealthPort)

		// Mark ready once consumer starts
		defer checker.Ready.Store(false)
		checker.Ready.Store(true)
	}

	if err := consumer.Run(ctx, cfg, db); err != nil {
		slog.Error("consumer error", "error", err.Error())
		os.Exit(1)
	}
}

func flagOrEnv(fs *flag.FlagSet, name, envKey, defaultVal, usage string) *string {
	val := fs.String(name, defaultVal, usage)
	if envVal := os.Getenv(envKey); envVal != "" && !isFlagSet(fs, name) {
		*val = envVal
	}
	return val
}

func isFlagSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
