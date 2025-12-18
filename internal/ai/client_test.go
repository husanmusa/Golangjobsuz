package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClientComplete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"hello"}`))
	}))
	t.Cleanup(ts.Close)

	client := NewHTTPClient(ts.URL)
	resp, err := client.Complete(context.Background(), "hi")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "hello" {
		t.Fatalf("unexpected response: %s", resp)
	}
}

func TestMockClientComplete(t *testing.T) {
	mock := &MockClient{Responses: map[string]string{"prompt": "value"}}
	resp, err := mock.Complete(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp != "value" {
		t.Fatalf("unexpected response: %s", resp)
	}
}
