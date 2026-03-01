package health

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"sync/atomic"
)

// Checker provides health and readiness endpoints.
type Checker struct {
	DB    *sql.DB
	Ready atomic.Bool
}

// Start starts the health HTTP server on the given port.
func (c *Checker) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", c.handleHealthz)
	mux.HandleFunc("/readyz", c.handleReadyz)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	go http.Serve(listener, mux)
	return nil
}

func (c *Checker) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (c *Checker) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !c.Ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"})
		return
	}

	if err := c.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
