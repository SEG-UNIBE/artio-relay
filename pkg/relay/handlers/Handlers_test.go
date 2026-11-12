package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip42"
)

const testServiceURL = "wss://relay.example.com"

/*
TestUnknownTypeHandler tests that unknown types are reported back
*/
func TestUnknownTypeHandler(t *testing.T) {
	handler := UnknownTypeHandler{Req: newRequest(t, "FOO", "bar")}

	if notice := handler.Handle(); notice != "the type FOO is unknown" {
		t.Fatalf("unexpected notice: %v", notice)
	}
}

/*
TestCloseHandler_IDRequired tests that CLOSE without a subscription id is rejected
*/
func TestCloseHandler_IDRequired(t *testing.T) {
	handler := CloseHandler{Req: newRequest(t, "CLOSE", "")}

	if notice := handler.Handle(); notice != "REQ has no <id>" {
		t.Fatalf("unexpected notice: %v", notice)
	}
}

/*
TestCloseHandler_Acknowledges tests that CLOSE with a subscription id is answered with a NOTICE
*/
func TestCloseHandler_Acknowledges(t *testing.T) {
	ws, client := newTestSocket(t, "", "")
	handler := CloseHandler{Ws: ws, Req: newRequest(t, "CLOSE", "sub")}

	if notice := handler.Handle(); notice != "" {
		t.Fatalf("unexpected notice: %v", notice)
	}

	envelope := readEnvelope(t, client)
	var label string
	_ = json.Unmarshal(envelope[0], &label)
	if label != "NOTICE" {
		t.Fatalf("expected a NOTICE but got %s", envelope)
	}
}

/*
TestCountHandler_FilterDecodeError tests that invalid filters are rejected
*/
func TestCountHandler_FilterDecodeError(t *testing.T) {
	handler := CountHandler{Req: newRequest(t, "COUNT", "sub", "not a filter")}

	if notice := handler.Handle(); notice != "failed to decode filter" {
		t.Fatalf("unexpected notice: %v", notice)
	}
}

/*
TestRequestHandler_IDNotString tests that a subscription id that is not a string is rejected
*/
func TestRequestHandler_IDNotString(t *testing.T) {
	handler := RequestHandler{Req: newRequest(t, "REQ", 42)}

	if notice := handler.Handle(); !strings.HasPrefix(notice, "failed to decode request: ") {
		t.Fatalf("unexpected notice: %v", notice)
	}
}

/*
TestEventHandler_InvalidID tests that events with a wrong id are rejected with an OK false
*/
func TestEventHandler_InvalidID(t *testing.T) {
	ws, client := newTestSocket(t, "", "")
	evt, _ := newSignedEvent(t, 1, "hello")
	evt.Content = "tampered"

	handler := EventHandler{Ws: ws, Req: newRequest(t, "EVENT", evt)}
	if notice := handler.Handle(); notice != "" {
		t.Fatalf("unexpected notice: %v", notice)
	}

	eventID, ok, reason := readOK(t, client)
	if eventID != evt.ID || ok || reason != "invalid: event id is computed incorrectly" {
		t.Fatalf("unexpected OK envelope: %v %v %v", eventID, ok, reason)
	}
}

/*
TestEventHandler_InvalidSignature tests that events with a signature of another event are rejected with an OK false
*/
func TestEventHandler_InvalidSignature(t *testing.T) {
	ws, client := newTestSocket(t, "", "")
	evt, sk := newSignedEvent(t, 1, "hello")

	// a valid signature by the same key, but for different content
	other := evt
	other.Content = "other"
	if err := other.Sign(sk); err != nil {
		t.Fatalf("failed to sign event: %v", err)
	}
	evt.Sig = other.Sig

	handler := EventHandler{Ws: ws, Req: newRequest(t, "EVENT", evt)}
	if notice := handler.Handle(); notice != "" {
		t.Fatalf("unexpected notice: %v", notice)
	}

	eventID, ok, reason := readOK(t, client)
	if eventID != evt.ID || ok || reason != "invalid: signature is invalid" {
		t.Fatalf("unexpected OK envelope: %v %v %v", eventID, ok, reason)
	}
}

/*
TestAuthHandler_Disabled tests that AUTH is ignored when NIP-42 is not supported
*/
func TestAuthHandler_Disabled(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.SupportNIP42 = false })
	ws, client := newTestSocket(t, "challenge", testServiceURL)
	ctx := context.Background()

	handler := AuthenticationHandler{Ctx: &ctx, Ws: ws, Req: newRequest(t, "AUTH", "anything")}
	if notice := handler.Handle(); notice != "" {
		t.Fatalf("unexpected notice: %v", notice)
	}
	expectNoMessage(t, client)
}

/*
TestAuthHandler_DecodeError tests that an invalid auth event is reported
*/
func TestAuthHandler_DecodeError(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.SupportNIP42 = true })
	ws, _ := newTestSocket(t, "challenge", testServiceURL)
	ctx := context.Background()

	handler := AuthenticationHandler{Ctx: &ctx, Ws: ws, Req: newRequest(t, "AUTH", "not an event")}
	if notice := handler.Handle(); !strings.HasPrefix(notice, "failed to decode auth event: ") {
		t.Fatalf("unexpected notice: %v", notice)
	}
}

/*
TestAuthHandler_Valid tests that a valid auth event authenticates the websocket
*/
func TestAuthHandler_Valid(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.SupportNIP42 = true })
	ws, client := newTestSocket(t, "challenge", testServiceURL)
	ctx := context.Background()

	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)
	evt := nip42.CreateUnsignedAuthEvent("challenge", pk, testServiceURL)
	if err := evt.Sign(sk); err != nil {
		t.Fatalf("failed to sign event: %v", err)
	}

	handler := AuthenticationHandler{Ctx: &ctx, Ws: ws, Req: newRequest(t, "AUTH", evt)}
	if notice := handler.Handle(); notice != "" {
		t.Fatalf("unexpected notice: %v", notice)
	}

	eventID, ok, _ := readOK(t, client)
	if eventID != evt.ID || !ok {
		t.Fatalf("expected a successful authentication but got %v %v", eventID, ok)
	}
	if ws.Authenticated != pk {
		t.Fatalf("websocket should be authenticated as %v but is %v", pk, ws.Authenticated)
	}
}

/*
TestAuthHandler_WrongChallenge tests that an auth event for another challenge is rejected
*/
func TestAuthHandler_WrongChallenge(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.SupportNIP42 = true })
	ws, client := newTestSocket(t, "challenge", testServiceURL)
	ctx := context.Background()

	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)
	evt := nip42.CreateUnsignedAuthEvent("another challenge", pk, testServiceURL)
	_ = evt.Sign(sk)

	handler := AuthenticationHandler{Ctx: &ctx, Ws: ws, Req: newRequest(t, "AUTH", evt)}
	_ = handler.Handle()

	_, ok, reason := readOK(t, client)
	if ok || reason != "error: failed to authenticate" {
		t.Fatalf("expected a failed authentication but got %v %v", ok, reason)
	}
	if ws.Authenticated != "" {
		t.Fatalf("websocket should not be authenticated but is %v", ws.Authenticated)
	}
}

/*
TestAuthHandler_WrongRelay tests that an auth event for another relay is rejected
*/
func TestAuthHandler_WrongRelay(t *testing.T) {
	useConfig(t, func(c *config.Configuration) { c.SupportNIP42 = true })
	ws, client := newTestSocket(t, "challenge", testServiceURL)
	ctx := context.Background()

	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)
	evt := nip42.CreateUnsignedAuthEvent("challenge", pk, "wss://other.example.com")
	_ = evt.Sign(sk)

	handler := AuthenticationHandler{Ctx: &ctx, Ws: ws, Req: newRequest(t, "AUTH", evt)}
	_ = handler.Handle()

	if _, ok, _ := readOK(t, client); ok {
		t.Fatalf("authentication for another relay should fail")
	}
}
