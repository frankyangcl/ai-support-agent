package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeGracefulShutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_ = listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	srv := &http.Server{Addr: listener.Addr().String(), Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })}
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, srv, time.Second) }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown timed out")
	}
}
