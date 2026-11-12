package webSocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/fasthttp/websocket"
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
newConnPair creates a connected websocket pair, the first one is the server side
*/
func newConnPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	serverConns := make(chan *websocket.Conn, 1)
	upgrader := NewUpgrader()
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
	return server, client
}

/*
TestNewUpgrader tests that the upgrader accepts every origin
*/
func TestNewUpgrader(t *testing.T) {
	upgrader := NewUpgrader()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://somewhere.else")

	if !upgrader.CheckOrigin(r) {
		t.Fatalf("the upgrader should accept every origin")
	}
}

/*
TestNewWebSocketFields tests that the constructor sets the given fields
*/
func TestNewWebSocketFields(t *testing.T) {
	ws := NewWebSocket(nil, "challenge", "wss://relay.example.com")

	if ws.Challenge != "challenge" {
		t.Fatalf("challenge not set, got %v", ws.Challenge)
	}
	if ws.ServiceURL != "wss://relay.example.com" {
		t.Fatalf("service url not set, got %v", ws.ServiceURL)
	}
	if ws.Authenticated != "" {
		t.Fatalf("a new websocket should not be authenticated, got %v", ws.Authenticated)
	}
}

/*
TestNewWebSocketNoLimiterWithoutConfig tests that no limiter is created when no rate is configured
*/
func TestNewWebSocketNoLimiterWithoutConfig(t *testing.T) {
	useConfig(t, func(c *config.Configuration) {
		c.RelayMessageRate = 0
		c.RelayMessageBurst = 0
	})

	ws := NewWebSocket(nil, "", "")
	if ws.limiter != nil {
		t.Fatalf("expected no limiter without a configured rate")
	}
	if err := ws.WaitForSlot(context.Background()); err != nil {
		t.Fatalf("WaitForSlot without limiter should never fail, got %v", err)
	}
}

/*
TestWaitForSlotRateLimited tests that messages beyond the burst have to wait for the rate
*/
func TestWaitForSlotRateLimited(t *testing.T) {
	useConfig(t, func(c *config.Configuration) {
		c.RelayMessageRate = 1
		c.RelayMessageBurst = 1
	})

	ws := NewWebSocket(nil, "", "")
	if ws.limiter == nil {
		t.Fatalf("expected a limiter with a configured rate")
	}

	if err := ws.WaitForSlot(context.Background()); err != nil {
		t.Fatalf("the first message should pass immediately, got %v", err)
	}

	// the next slot is only available after one second, which is longer than the context allows
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := ws.WaitForSlot(ctx); err == nil {
		t.Fatalf("the second message should be rate limited")
	}
}

/*
TestWaitForSlotCancelledContext tests that a cancelled connection stops waiting
*/
func TestWaitForSlotCancelledContext(t *testing.T) {
	useConfig(t, func(c *config.Configuration) {
		c.RelayMessageRate = 1
		c.RelayMessageBurst = 1
	})

	ws := NewWebSocket(nil, "", "")
	_ = ws.WaitForSlot(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ws.WaitForSlot(ctx); err == nil {
		t.Fatalf("WaitForSlot should fail on a cancelled context")
	}
}

/*
TestWriteJSON tests that JSON written to the websocket arrives at the client
*/
func TestWriteJSON(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.RelayWriteWait = time.Second })
	server, client := newConnPair(t)
	ws := NewWebSocket(server, "", "")

	if err := ws.WriteJSON([]string{"NOTICE", "hello"}); err != nil {
		t.Fatalf("failed to write JSON: %v", err)
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	var received []string
	if err := client.ReadJSON(&received); err != nil {
		t.Fatalf("failed to read JSON: %v", err)
	}
	if len(received) != 2 || received[0] != "NOTICE" || received[1] != "hello" {
		t.Fatalf("unexpected message: %v", received)
	}
}

/*
TestWriteMessage tests that raw messages written to the websocket arrive at the client
*/
func TestWriteMessage(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.RelayWriteWait = time.Second })
	server, client := newConnPair(t)
	ws := NewWebSocket(server, "", "")

	if err := ws.WriteMessage(websocket.TextMessage, []byte("raw")); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	typ, message, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if typ != websocket.TextMessage || string(message) != "raw" {
		t.Fatalf("unexpected message %v of type %v", string(message), typ)
	}
}

/*
TestWriteJSONConcurrent tests that concurrent writers do not corrupt messages
*/
func TestWriteJSONConcurrent(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.RelayWriteWait = 5 * time.Second })
	server, client := newConnPair(t)
	ws := NewWebSocket(server, "", "")

	const writers = 50
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = ws.WriteJSON(map[string]int{"n": i})
		}(i)
	}

	seen := make(map[int]bool)
	_ = client.SetReadDeadline(time.Now().Add(5 * time.Second))
	for i := 0; i < writers; i++ {
		_, message, err := client.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read message %v: %v", i, err)
		}
		var payload map[string]int
		if err := json.Unmarshal(message, &payload); err != nil {
			t.Fatalf("received corrupted message %q: %v", message, err)
		}
		seen[payload["n"]] = true
	}
	wg.Wait()

	if len(seen) != writers {
		t.Fatalf("expected %v distinct messages but got %v", writers, len(seen))
	}
}

/*
TestGetRemoteIP tests that the remote address of the connection is returned
*/
func TestGetRemoteIP(t *testing.T) {
	server, client := newConnPair(t)
	ws := NewWebSocket(server, "", "")

	if ip := ws.GetRemoteIP(); ip != client.LocalAddr().String() {
		t.Fatalf("expected %v but got %v", client.LocalAddr().String(), ip)
	}
}
