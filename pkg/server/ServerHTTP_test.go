package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/relay"
	"github.com/fasthttp/websocket"
	"github.com/nbd-wtf/go-nostr"
)

/*
useConfig changes the global config for a single test and restores it afterwards
*/
func useConfig(t *testing.T, change func(c *config.Configuration)) {
	t.Helper()
	original := *config.Config
	t.Cleanup(func() { *config.Config = original })
	change(config.Config)
}

/*
useConnectionConfig sets the timings and limits needed to serve websocket connections in tests
*/
func useConnectionConfig(t *testing.T, change func(c *config.Configuration)) {
	t.Helper()
	useConfig(t, func(c *config.Configuration) {
		c.RelayWriteWait = 2 * time.Second
		c.RelayPongWait = 10 * time.Second
		c.RelayPingWait = 5 * time.Second
		c.RelayMaxMessageSize = 64 * 1024
		c.RelayMessageRate = 100
		c.RelayMessageBurst = 100
		c.SupportNIP42 = false
		if change != nil {
			change(c)
		}
	})
}

/*
dialTestServer starts the relay server on a test listener and connects a websocket client
*/
func dialTestServer(t *testing.T) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(NewServer(&relay.Relay{}))
	t.Cleanup(srv.Close)

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

/*
readMessage reads the next message the client receives
*/
func readMessage(t *testing.T, client *websocket.Conn) string {
	t.Helper()
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	return string(message)
}

/*
TestServeHTTP_NIP11 tests that the relay information document is served for application/nostr+json
*/
func TestServeHTTP_NIP11(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.NIP11Description = "test relay" })
	s := NewServer(&relay.Relay{Name: "artio"})

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "application/nostr+json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 but got %v", w.Code)
	}
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json but got %v", contentType)
	}

	var info map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if info["name"] != "artio" || info["description"] != "test relay" {
		t.Fatalf("unexpected relay information: %v", info)
	}
	if _, ok := info["supported_nips"]; !ok {
		t.Fatalf("supported_nips missing in %v", info)
	}
}

/*
TestServeHTTP_UnknownPath tests that other requests are delegated to the serve mux
*/
func TestServeHTTP_UnknownPath(t *testing.T) {
	s := NewServer(&relay.Relay{})

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/does-not-exist", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 but got %v", w.Code)
	}
}

/*
TestInjectHandler tests that injected handlers are served on their path
*/
func TestInjectHandler(t *testing.T) {
	s := NewServer(&relay.Relay{})
	s.InjectHandler("/custom", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/custom", nil))

	if w.Code != http.StatusTeapot {
		t.Fatalf("expected status 418 but got %v", w.Code)
	}
}

/*
TestServeHTTP_FailedUpgrade tests that a broken websocket handshake is rejected without stopping the server
*/
func TestServeHTTP_FailedUpgrade(t *testing.T) {
	useConnectionConfig(t, nil)
	s := NewServer(&relay.Relay{})

	// only the Upgrade header, the rest of the handshake is missing
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Upgrade", "websocket")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 but got %v", w.Code)
	}
}

/*
TestWebsocket_UnknownType tests a full round trip through the server for an unknown request type
*/
func TestWebsocket_UnknownType(t *testing.T) {
	useConnectionConfig(t, nil)
	client := dialTestServer(t)

	if err := client.WriteMessage(websocket.TextMessage, []byte(`["FOO", "bar"]`)); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if message := readMessage(t, client); !strings.Contains(message, "the type FOO is unknown") {
		t.Fatalf("unexpected message: %v", message)
	}
}

/*
TestWebsocket_InvalidEvent tests a full round trip through the server for an event with a wrong id
*/
func TestWebsocket_InvalidEvent(t *testing.T) {
	useConnectionConfig(t, nil)
	client := dialTestServer(t)

	evt := nostr.Event{ID: strings.Repeat("0", 64), CreatedAt: nostr.Now(), Kind: 1, Tags: nostr.Tags{}}
	if err := client.WriteJSON([]any{"EVENT", evt}); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if message := readMessage(t, client); !strings.Contains(message, "invalid: event id is computed incorrectly") {
		t.Fatalf("unexpected message: %v", message)
	}
}

/*
TestWebsocket_AuthChallenge tests that the AUTH challenge is sent on connect when NIP-42 is enabled
*/
func TestWebsocket_AuthChallenge(t *testing.T) {
	useConnectionConfig(t, func(c *config.Configuration) { c.SupportNIP42 = true })
	client := dialTestServer(t)

	var envelope []string
	if err := json.Unmarshal([]byte(readMessage(t, client)), &envelope); err != nil {
		t.Fatalf("AUTH message is not a JSON array: %v", err)
	}
	if len(envelope) != 2 || envelope[0] != "AUTH" || len(envelope[1]) != 16 {
		t.Fatalf("expected an AUTH challenge but got %v", envelope)
	}
}

/*
TestWebsocket_MessageTooLarge tests that messages above the configured size close the connection
*/
func TestWebsocket_MessageTooLarge(t *testing.T) {
	useConnectionConfig(t, func(c *config.Configuration) { c.RelayMaxMessageSize = 128 })
	client := dialTestServer(t)

	large := `["FOO", "` + strings.Repeat("x", 1024) + `"]`
	_ = client.WriteMessage(websocket.TextMessage, []byte(large))

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, message, err := client.ReadMessage(); err == nil {
		t.Fatalf("expected the connection to be closed but got %s", message)
	} else if !websocket.IsCloseError(err, websocket.CloseMessageTooBig) && !strings.Contains(err.Error(), "close") {
		t.Fatalf("expected a close error but got %v", err)
	}
}
