package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestDeepSeekStreamChatUsesProviderStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request chatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if !request.Stream {
			t.Error("stream flag was false")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"one\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" two\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()
	client := NewDeepSeekClient("test-key")
	client.BaseURL = server.URL
	var deltas []string
	err := client.StreamChat(context.Background(), "system", "user", func(delta string) error { deltas = append(deltas, delta); return nil })
	if err != nil {
		t.Fatalf("StreamChat: %v", err)
	}
	if !reflect.DeepEqual(deltas, []string{"one", " two"}) {
		t.Fatalf("unexpected deltas %#v", deltas)
	}
}

func TestDeepSeekStreamChatPropagatesCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	client := NewDeepSeekClient("test-key")
	client.BaseURL = server.URL
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.StreamChat(ctx, "system", "user", func(string) error { return nil }); err == nil {
		t.Fatal("expected cancellation error")
	}
}
