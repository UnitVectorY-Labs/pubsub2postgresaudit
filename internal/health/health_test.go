package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz_ReturnsOK(t *testing.T) {
	c := &Checker{}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	c.handleHealthz(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q, want application/json", ct)
	}
}

func TestReadyz_NotReady(t *testing.T) {
	c := &Checker{}
	// Ready is false by default, and DB is nil — should be not ready
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	c.handleReadyz(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}
