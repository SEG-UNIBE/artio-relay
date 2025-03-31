package handlers

import (
	"artio-relay/pkg/storage/adapter"
	"artio-relay/pkg/webSocket"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"log"
)

/*
EventHandler handles messages of type EVENT
*/
type EventHandler struct {
	Ctx any
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

/*
Handle handles the functionality for this event
*/
func (e EventHandler) Handle() string {

	latestIndex := len(e.Req) - 1
	// we have received a new EVENT from the client that we need to handle

	// it's a new event
	var evt nostr.Event
	if err := json.Unmarshal(e.Req[latestIndex], &evt); err != nil {
		return "failed to decode event: " + err.Error()
	}

	// check id and return error if its NOK
	hash := sha256.Sum256(evt.Serialize())
	if id := hex.EncodeToString(hash[:]); id != evt.ID {
		reason := "invalid: event id is computed incorrectly"
		return fmt.Sprintf("%v", nostr.OKEnvelope{EventID: evt.ID, OK: false, Reason: reason})
	}

	// check signature
	if ok, err := evt.CheckSignature(); err != nil {
		return fmt.Sprintf("%v", nostr.OKEnvelope{EventID: evt.ID, OK: false, Reason: "error: failed to verify signature"})
	} else if !ok {
		return fmt.Sprintf("%v", nostr.OKEnvelope{EventID: evt.ID, OK: false, Reason: "invalid: signature is invalid"})
	}

	eventAdapter := adapter.EventAdapter{}
	_, err := eventAdapter.Create(evt)
	if err != nil {
		log.Printf("Error occured %v", err)
	}
	return fmt.Sprintf("%v", nostr.OKEnvelope{EventID: evt.ID, OK: true, Reason: ""})

}
