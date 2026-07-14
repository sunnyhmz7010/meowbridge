package meow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPushSendsJSONToNicknamePath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Query().Get("msgType") != "html" || r.URL.Query().Get("htmlHeight") != "500" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := New(server.URL, time.Second)
	resp, err := client.Push(context.Background(), PushRequest{
		Nickname:   "sunny",
		Title:      "title",
		Msg:        "message",
		MsgType:    "html",
		HTMLHeight: 500,
	})
	if err != nil {
		t.Fatalf("Push: %v", err)
	}
	if gotPath != "/sunny" {
		t.Fatalf("path = %q", gotPath)
	}
	if resp.StatusCode != http.StatusOK || resp.Body != `{"ok":true}` {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestPushTreatsNon2xxAsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}))
	defer server.Close()

	client := New(server.URL, time.Second)
	resp, err := client.Push(context.Background(), PushRequest{Nickname: "sunny", Msg: "message", MsgType: "text"})
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if resp.StatusCode != http.StatusBadGateway || resp.Body != "bad gateway" {
		t.Fatalf("resp = %#v", resp)
	}
}
