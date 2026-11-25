package relay

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage"
	"github.com/SEG-UNIBE/artio-relay/pkg/webSocket"
	"github.com/fasthttp/websocket"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip42"
)

const testServiceURL = "wss://relay.example.com"

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
newTestSocket creates a relay side websocket using the relay challenge, connected to a test client
*/
func newTestSocket(t *testing.T, relay *Relay) (*webSocket.WebSocket, *websocket.Conn) {
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
	return relay.Challenge(server), client
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
TestNewRelay tests that the relay is created with the configured name
*/
func TestNewRelay(t *testing.T) {
	r := NewRelay(storage.Storage{})

	if r.Name != config.Config.RelayName() {
		t.Fatalf("relay name should be %v but got %v", config.Config.RelayName(), r.Name)
	}
}

/*
TestServiceURL tests that the service url is taken from the config
*/
func TestServiceURL(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.RelayServiceURL = testServiceURL })
	r := Relay{}

	if url := r.ServiceURL(); url != testServiceURL {
		t.Fatalf("expected %v but got %v", testServiceURL, url)
	}
}

/*
TestChallengeFormat tests that challenges are 8 random bytes in hex and differ per connection
*/
func TestChallengeFormat(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.RelayServiceURL = testServiceURL })
	r := Relay{}

	first := r.Challenge(nil)
	second := r.Challenge(nil)

	if decoded, err := hex.DecodeString(first.Challenge); err != nil || len(decoded) != 8 {
		t.Fatalf("challenge should be 8 hex encoded bytes but got %v", first.Challenge)
	}
	if first.Challenge == second.Challenge {
		t.Fatalf("two challenges should not be equal: %v", first.Challenge)
	}
	if first.ServiceURL != testServiceURL {
		t.Fatalf("service url should be %v but got %v", testServiceURL, first.ServiceURL)
	}
}

/*
TestNIP11SupportedNIPs tests that NIP-42 is only advertised when it is enabled
*/
func TestNIP11SupportedNIPs(t *testing.T) {
	r := Relay{}

	useConfig(t, func(c *config.Configuration) { c.SupportNIP42 = false })
	if slices.Contains(r.GetNIP11Information().SupportedNIPs, any(42)) {
		t.Fatalf("NIP-42 should not be advertised when disabled")
	}

	config.Config.SupportNIP42 = true
	nips := r.GetNIP11Information().SupportedNIPs
	for _, nip := range []int{1, 2, 9, 11, 42, 45, 50, 65} {
		if !slices.Contains(nips, any(nip)) {
			t.Fatalf("NIP-%v should be advertised, got %v", nip, nips)
		}
	}
}

/*
TestNIP11Fields tests that the NIP-11 document is filled from the config
*/
func TestNIP11Fields(t *testing.T) {
	useConfig(t, func(c *config.Configuration) {
		c.NIP11Description = "description"
		c.NIP11Pubkey = "pubkey"
		c.NIP11Contact = "contact"
		c.NIP11Software = "software"
		c.NIP11Version = "version"
		c.NIP11Banner = "banner"
	})
	r := Relay{Name: "name"}

	info := r.GetNIP11Information()
	if info.Name != "name" || info.Description != "description" || info.PubKey != "pubkey" ||
		info.Contact != "contact" || info.Software != "software" || info.Version != "version" || info.Banner != "banner" {
		t.Fatalf("NIP-11 document does not match the config: %+v", info)
	}
}

/*
TestNIP11Icon tests that the NIP-11 icon is taken from the icon setting
*/
func TestNIP11Icon(t *testing.T) {

	useConfig(t, func(c *config.Configuration) {
		c.NIP11Icon = "icon"
		c.NIP11Banner = "banner"
	})
	r := Relay{}

	if icon := r.GetNIP11Information().Icon; icon != "icon" {
		t.Fatalf("icon should be icon but got %v", icon)
	}
}

/*
TestSendAuthMessage tests that the AUTH challenge is only sent when NIP-42 is enabled
*/
func TestSendAuthMessage(t *testing.T) {
	r := Relay{}

	useConfig(t, func(c *config.Configuration) { c.SupportNIP42 = false })
	ws, client := newTestSocket(t, &r)
	r.SendAuthMessage(ws)
	expectNoMessage(t, client)

	config.Config.SupportNIP42 = true
	ws, client = newTestSocket(t, &r)
	r.SendAuthMessage(ws)

	var envelope []string
	if err := json.Unmarshal([]byte(readMessage(t, client)), &envelope); err != nil {
		t.Fatalf("AUTH message is not a JSON array: %v", err)
	}
	if len(envelope) != 2 || envelope[0] != "AUTH" || envelope[1] != ws.Challenge {
		t.Fatalf("expected [AUTH, %v] but got %v", ws.Challenge, envelope)
	}
}

/*
TestHandleMessage_InvalidJSON tests that invalid JSON is silently ignored
*/
func TestHandleMessage_InvalidJSON(t *testing.T) {
	r := Relay{}
	ws, client := newTestSocket(t, &r)
	ctx := context.Background()

	r.HandleMessage(&ctx, ws, []byte("not json"))
	expectNoMessage(t, client)
}

/*
TestHandleMessage_TooShort tests that requests with less than 2 elements are reported
*/
func TestHandleMessage_TooShort(t *testing.T) {
	r := Relay{}
	ws, client := newTestSocket(t, &r)
	ctx := context.Background()

	r.HandleMessage(&ctx, ws, []byte(`["REQ"]`))
	if message := readMessage(t, client); !strings.Contains(message, "request has less than 2 parameters") {
		t.Fatalf("unexpected message: %v", message)
	}
}

/*
TestHandleMessage_UnknownType tests that unknown request types are reported
*/
func TestHandleMessage_UnknownType(t *testing.T) {
	r := Relay{}
	ws, client := newTestSocket(t, &r)
	ctx := context.Background()

	r.HandleMessage(&ctx, ws, []byte(`["FOO", "bar"]`))
	if message := readMessage(t, client); !strings.Contains(message, "the type FOO is unknown") {
		t.Fatalf("unexpected message: %v", message)
	}
}

/*
TestHandleMessage_Close tests that CLOSE is dispatched to the close handler
*/
func TestHandleMessage_Close(t *testing.T) {
	r := Relay{}
	ws, client := newTestSocket(t, &r)
	ctx := context.Background()

	r.HandleMessage(&ctx, ws, []byte(`["CLOSE", "sub"]`))
	if message := readMessage(t, client); !strings.HasPrefix(message, `["NOTICE"`) {
		t.Fatalf("expected a NOTICE but got %v", message)
	}
}

/*
TestHandleMessage_EventInvalidID tests that EVENT is dispatched to the event handler
*/
func TestHandleMessage_EventInvalidID(t *testing.T) {
	r := Relay{}
	ws, client := newTestSocket(t, &r)
	ctx := context.Background()

	evt := nostr.Event{ID: strings.Repeat("0", 64), CreatedAt: nostr.Now(), Kind: 1, Tags: nostr.Tags{}}
	request, _ := json.Marshal([]any{"EVENT", evt})

	r.HandleMessage(&ctx, ws, request)
	if message := readMessage(t, client); !strings.Contains(message, "invalid: event id is computed incorrectly") {
		t.Fatalf("unexpected message: %v", message)
	}
}

/*
TestHandleMessage_RecoversFromPanic tests that a panicking handler does not crash the relay.
A nil context makes the AUTH handler panic after a successful validation.
*/
func TestHandleMessage_RecoversFromPanic(t *testing.T) {
	useConfig(t, func(c *config.Configuration) {
		c.SupportNIP42 = true
		c.RelayServiceURL = testServiceURL
	})
	r := Relay{}
	ws, _ := newTestSocket(t, &r)

	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)
	evt := nip42.CreateUnsignedAuthEvent(ws.Challenge, pk, testServiceURL)
	_ = evt.Sign(sk)
	request, _ := json.Marshal([]any{"AUTH", evt})

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("HandleMessage should recover from panics but got: %v", recovered)
		}
	}()
	r.HandleMessage(nil, ws, request)
}
