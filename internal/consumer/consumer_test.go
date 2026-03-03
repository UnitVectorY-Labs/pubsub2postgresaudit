package consumer

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestParseSubscription_Valid(t *testing.T) {
	proj, sub, err := ParseSubscription("projects/my-project/subscriptions/my-sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj != "my-project" {
		t.Errorf("project = %q, want %q", proj, "my-project")
	}
	if sub != "my-sub" {
		t.Errorf("subscription = %q, want %q", sub, "my-sub")
	}
}

func TestParseSubscription_Invalid(t *testing.T) {
	cases := []string{
		"",
		"my-sub",
		"projects/my-project",
		"projects/my-project/subscriptions",
		"projects/my-project/topics/my-topic",
		"projects//subscriptions/",
		"a/b/c/d",
	}
	for _, tc := range cases {
		_, _, err := ParseSubscription(tc)
		if err == nil {
			t.Errorf("ParseSubscription(%q) expected error, got nil", tc)
		}
	}
}

func compressGzip(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func compressZstd(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatalf("zstd writer: %v", err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("zstd write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zstd close: %v", err)
	}
	return buf.Bytes()
}

func TestDecompressData_Gzip(t *testing.T) {
	original := []byte(`{"key":"value"}`)
	compressed := compressGzip(t, original)

	out, err := decompressData("gzip", compressed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(out, original) {
		t.Errorf("got %q, want %q", out, original)
	}
}

func TestDecompressData_Zstd(t *testing.T) {
	original := []byte(`{"key":"value"}`)
	compressed := compressZstd(t, original)

	out, err := decompressData("zstd", compressed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(out, original) {
		t.Errorf("got %q, want %q", out, original)
	}
}

func TestDecompressData_InvalidAlgorithm(t *testing.T) {
	_, err := decompressData("brotli", []byte("data"))
	if err == nil {
		t.Error("expected error for unsupported algorithm, got nil")
	}
}

func TestDecompressData_GzipInvalidData(t *testing.T) {
	_, err := decompressData("gzip", []byte("not gzip data"))
	if err == nil {
		t.Error("expected error for invalid gzip data, got nil")
	}
}

func TestDecompressData_ZstdInvalidData(t *testing.T) {
	_, err := decompressData("zstd", []byte("not zstd data"))
	if err == nil {
		t.Error("expected error for invalid zstd data, got nil")
	}
}
