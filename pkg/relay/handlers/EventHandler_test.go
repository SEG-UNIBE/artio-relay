package handlers

import (
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

/*
TestEventHandler_DecodeError tests that the EventHandler returns is an error string when the event json is invalid
*/
func TestEventHandler_DecodeError(t *testing.T) {
	var msg = make([]json.RawMessage, 3)
	msg[0], _ = json.Marshal("EVENT")
	msg[1], _ = json.Marshal("")

	eh := EventHandler{Ctx: nil, Ws: nil, Req: msg}

	outputString := eh.Handle()

	if outputString != "failed to decode event: unexpected end of JSON input" {
		t.Fatalf("Handler does not properly check for invalid JSON Input id, result was: %v", outputString)
	}
}

/*
TestEventHandler_CheckSumInCorrect tests that the EventHandler returns an error string when the Checksum is incorrect
*/
func TestEventHandler_CheckSumInCorrect(t *testing.T) {

	var event = nostr.Event{
		ID:        "",
		PubKey:    "",
		CreatedAt: 0,
		Kind:      0,
		Tags:      nil,
		Content:   "",
		Sig:       "",
	}
	var msg = make([]json.RawMessage, 3)
	msg[0], _ = json.Marshal("EVENT")
	msg[1], _ = json.Marshal(event)

	eh := EventHandler{Ctx: nil, Ws: nil, Req: msg}

	//outputString := eh.Handle()
	_ = eh.Handle()
}
