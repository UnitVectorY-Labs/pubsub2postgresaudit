package consumer

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/klauspost/compress/zstd"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/config"
	"github.com/UnitVectorY-Labs/pubsub2postgresaudit/internal/database"
)

// TestIntegration_CompressionEndToEnd tests the full consumer pipeline with
// real PostgreSQL and Pub/Sub emulator containers covering all compression
// scenarios: no compression, gzip, zstd, and failure modes.
func TestIntegration_CompressionEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	// ── PostgreSQL container ────────────────────────────────────────────────
	pgContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image: "postgres:18",
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "test",
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("starting postgres container: %v", err)
	}
	tc.CleanupContainer(t, pgContainer)

	pgHost, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("getting postgres host: %v", err)
	}
	pgPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("getting postgres port: %v", err)
	}

	// ── Pub/Sub emulator container ──────────────────────────────────────────
	pubsubContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image: "gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators",
			Cmd: []string{
				"gcloud", "beta", "emulators", "pubsub", "start",
				"--host-port=0.0.0.0:8085",
				"--project=test-project",
			},
			ExposedPorts: []string{"8085/tcp"},
			WaitingFor:   wait.ForListeningPort("8085/tcp").WithStartupTimeout(120 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("starting pubsub emulator container: %v", err)
	}
	tc.CleanupContainer(t, pubsubContainer)

	psHost, err := pubsubContainer.Host(ctx)
	if err != nil {
		t.Fatalf("getting pubsub host: %v", err)
	}
	psPort, err := pubsubContainer.MappedPort(ctx, "8085/tcp")
	if err != nil {
		t.Fatalf("getting pubsub port: %v", err)
	}
	t.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("%s:%s", psHost, psPort.Port()))

	// ── Database setup ──────────────────────────────────────────────────────
	connStr := fmt.Sprintf("host=%s port=%s user=test password=test dbname=test sslmode=disable",
		pgHost, pgPort.Port())
	db, err := database.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("connecting to postgres: %v", err)
	}
	defer db.Close()

	const schema, table = "public", "audit_log"
	if err := database.Migrate(ctx, db, schema, table); err != nil {
		t.Fatalf("migrating database: %v", err)
	}

	// ── Pub/Sub topic + subscription ────────────────────────────────────────
	const projectID = "test-project"
	psClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("creating pubsub client: %v", err)
	}
	defer psClient.Close()

	topic, err := psClient.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("creating topic: %v", err)
	}
	if _, err = psClient.CreateSubscription(ctx, "test-sub", pubsub.SubscriptionConfig{Topic: topic}); err != nil {
		t.Fatalf("creating subscription: %v", err)
	}

	// ── Consumer ────────────────────────────────────────────────────────────
	cfg := &config.Config{
		DBSchema:           schema,
		DBTable:            table,
		PubSubSubscription: fmt.Sprintf("projects/%s/subscriptions/test-sub", projectID),
	}
	consumerCtx, stopConsumer := context.WithCancel(ctx)
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- Run(consumerCtx, cfg, db)
	}()
	defer func() {
		stopConsumer()
		<-consumerDone
	}()

	// ── Helpers ─────────────────────────────────────────────────────────────
	gzipCompress := func(data []byte) []byte {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, _ = w.Write(data)
		_ = w.Close()
		return buf.Bytes()
	}
	zstdCompress := func(data []byte) []byte {
		var buf bytes.Buffer
		w, _ := zstd.NewWriter(&buf)
		_, _ = w.Write(data)
		_ = w.Close()
		return buf.Bytes()
	}
	publish := func(data []byte, attrs map[string]string) {
		t.Helper()
		res := topic.Publish(ctx, &pubsub.Message{Data: data, Attributes: attrs})
		if _, err := res.Get(ctx); err != nil {
			t.Fatalf("publishing message: %v", err)
		}
	}
	countRows := func(jsonFilter string) int {
		t.Helper()
		var n int
		query := fmt.Sprintf(`SELECT COUNT(*) FROM %q.%q WHERE data @> $1::jsonb`, schema, table)
		if err := db.QueryRowContext(ctx, query, jsonFilter).Scan(&n); err != nil {
			t.Fatalf("counting rows for filter %s: %v", jsonFilter, err)
		}
		return n
	}
	waitForTotal := func(want int) {
		t.Helper()
		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			var n int
			if err := db.QueryRowContext(ctx,
				fmt.Sprintf(`SELECT COUNT(*) FROM %q.%q`, schema, table),
			).Scan(&n); err == nil && n >= want {
				return
			}
			time.Sleep(250 * time.Millisecond)
		}
		t.Fatalf("timed out waiting for %d rows in database", want)
	}

	// ── Publish all test messages ────────────────────────────────────────────
	//
	// Good messages (should be inserted):
	//   1. Plain JSON – no compression attribute
	//   2. gzip-compressed JSON – compression=gzip
	//   3. zstd-compressed JSON – compression=zstd
	//
	// Failure modes (should NOT be inserted):
	//   4. gzip-compressed bytes WITHOUT compression attribute → invalid_json
	//   5. Data with an unknown compression attribute value   → decompression_error
	//
	// Sentinel (inserted last to confirm all previous messages were processed):
	//   6. Plain JSON sentinel

	publish([]byte(`{"scenario":"plain"}`), nil)
	publish(gzipCompress([]byte(`{"scenario":"gzip"}`)), map[string]string{"compression": "gzip"})
	publish(zstdCompress([]byte(`{"scenario":"zstd"}`)), map[string]string{"compression": "zstd"})
	publish(gzipCompress([]byte(`{"scenario":"missing-attr"}`)), nil)
	publish(gzipCompress([]byte(`{"scenario":"unknown-algo"}`)), map[string]string{"compression": "brotli"})
	publish([]byte(`{"scenario":"sentinel"}`), nil) // sentinel

	// Wait until the 4 expected rows are present (plain + gzip + zstd + sentinel).
	waitForTotal(4)

	// ── Assertions ───────────────────────────────────────────────────────────

	// Scenario 1: plain JSON inserted correctly.
	if n := countRows(`{"scenario":"plain"}`); n != 1 {
		t.Errorf("plain: want 1 row, got %d", n)
	}

	// Scenario 2: gzip-compressed JSON decompressed and inserted correctly.
	if n := countRows(`{"scenario":"gzip"}`); n != 1 {
		t.Errorf("gzip: want 1 row, got %d", n)
	}

	// Scenario 3: zstd-compressed JSON decompressed and inserted correctly.
	if n := countRows(`{"scenario":"zstd"}`); n != 1 {
		t.Errorf("zstd: want 1 row, got %d", n)
	}

	// Sentinel inserted (proves the consumer has processed all preceding messages).
	if n := countRows(`{"scenario":"sentinel"}`); n != 1 {
		t.Errorf("sentinel: want 1 row, got %d", n)
	}

	// Failure mode 4: gzip bytes without compression attribute → NOT inserted.
	if n := countRows(`{"scenario":"missing-attr"}`); n != 0 {
		t.Errorf("missing-attr: want 0 rows, got %d", n)
	}

	// Total row count must be exactly 4.
	var total int
	if err := db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM %q.%q`, schema, table),
	).Scan(&total); err != nil {
		t.Fatalf("counting total rows: %v", err)
	}
	if total != 4 {
		t.Errorf("total rows: want 4, got %d", total)
	}
}
