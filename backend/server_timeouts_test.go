package main

import (
	"net/http"
	"testing"
	"time"
)

// The server must never go back to a bare http.ListenAndServe: without these
// timeouts a slowloris client holds a connection for as long as it likes.
func TestNewServerSetsEveryTimeout(t *testing.T) {
	srv := newServer(":3001", http.NotFoundHandler())

	checks := map[string]time.Duration{
		"ReadHeaderTimeout": srv.ReadHeaderTimeout,
		"ReadTimeout":       srv.ReadTimeout,
		"WriteTimeout":      srv.WriteTimeout,
		"IdleTimeout":       srv.IdleTimeout,
	}
	for name, got := range checks {
		if got <= 0 {
			t.Errorf("%s is not set (%v)", name, got)
		}
	}
	if srv.MaxHeaderBytes <= 0 {
		t.Errorf("MaxHeaderBytes is not set (%d)", srv.MaxHeaderBytes)
	}
}

// /api/proxy allows its outbound client 15s. A WriteTimeout at or below that
// would cut off a request the proxy is still legitimately serving.
func TestWriteTimeoutOutlastsTheSlowestHandler(t *testing.T) {
	const slowestHandler = 15 * time.Second

	srv := newServer(":3001", http.NotFoundHandler())
	if srv.WriteTimeout <= slowestHandler {
		t.Fatalf("WriteTimeout = %v, must be longer than %v", srv.WriteTimeout, slowestHandler)
	}
}
