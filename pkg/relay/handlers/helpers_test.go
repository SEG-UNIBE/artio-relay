package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/webSocket"
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
newTestSocket creates a relay side websocket connected to a test client
*/
func newTestSocket(t *testing.T, challenge string, serviceURL string) (*webSocket.WebSocket, *websocket.Conn) {
	t.Helper()
	serverConns := make(chan *websocket.Conn, 1)
	upgrader := webSocket.NewUpgrader()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("failed to upgrade: %v", err)
			return
		}
		serverConns <- conn
	}))
	t.Cleanup(srv.Close)

	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	server := <-serverConns
	t.Cleanup(func() { _ = server.Close() })
	return webSocket.NewWebSocket(server, challenge, serviceURL), client
}

/*
readEnvelope reads the next message from the client and decodes it into its raw parts
*/
func readEnvelope(t *testing.T, client *websocket.Conn) []json.RawMessage {
	t.Helper()
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	var envelope []json.RawMessage
	if err := json.Unmarshal(message, &envelope); err != nil {
		t.Fatalf("message is not a JSON array: %s", message)
	}
	return envelope
}

/*
readOK reads the next message and checks that it is an OK envelope
*/
func readOK(t *testing.T, client *websocket.Conn) (eventID string, ok bool, reason string) {
	t.Helper()
	envelope := readEnvelope(t, client)
	var label string
	_ = json.Unmarshal(envelope[0], &label)
	if label != "OK" || len(envelope) != 4 {
		t.Fatalf("expected an OK envelope but got %s", envelope)
	}
	_ = json.Unmarshal(envelope[1], &eventID)
	_ = json.Unmarshal(envelope[2], &ok)
	_ = json.Unmarshal(envelope[3], &reason)
	return eventID, ok, reason
}

/*
expectNoMessage checks that the client does not receive anything within a short time
*/
func expectNoMessage(t *testing.T, client *websocket.Conn) {
	t.Helper()
	_ = client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if _, message, err := client.ReadMessage(); err == nil {
		t.Fatalf("expected no message but got %s", message)
	}
}

/*
newRequest builds a raw relay request out of the given parts
*/
func newRequest(t *testing.T, parts ...any) []json.RawMessage {
	t.Helper()
	request := make([]json.RawMessage, len(parts))
	for i, part := range parts {
		raw, err := json.Marshal(part)
		if err != nil {
			t.Fatalf("failed to marshal request part: %v", err)
		}
		request[i] = raw
	}
	return request
}

/*
newSignedEvent creates an event signed by a new random key
*/
func newSignedEvent(t *testing.T, kind int, content string) (nostr.Event, string) {
	t.Helper()
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		t.Fatalf("failed to derive public key: %v", err)
	}
	evt := nostr.Event{PubKey: pk, CreatedAt: nostr.Now(), Kind: kind, Tags: nostr.Tags{}, Content: content}
	if err := evt.Sign(sk); err != nil {
		t.Fatalf("failed to sign event: %v", err)
	}
	return evt, sk
}
