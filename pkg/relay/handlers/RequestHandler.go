package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/adapter"
	"github.com/SEG-UNIBE/artio-relay/pkg/webSocket"

	"github.com/nbd-wtf/go-nostr"
)

/*
RequestHandler handles messages of type REQ
*/
type RequestHandler struct {
	Ctx *context.Context
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

/*
Handle handles the functionality for this event
*/
func (r RequestHandler) Handle() string {
	var id string
	err := json.Unmarshal(r.Req[1], &id)
	if err != nil {
		return "failed to decode request: " + err.Error()
	}
	if id == "" {
		return "REQ has no <id>"
	}

	filters := make(nostr.Filters, len(r.Req)-2)
	for i, filterReq := range r.Req[2:] {
		filters[i].Limit = config.Config.RelayMaxMessageCount
		if err := json.Unmarshal(
			filterReq,
			&filters[i],
		); err != nil {
			return "failed to decode filter"
		}
	}

	// we have fetched all the filters, so we can fetch them.

	for _, filter := range filters {

		eventAdapter := adapter.EventAdapter{}
		var events []nostr.Event
		events, err := eventAdapter.Get(filter)

		if err != nil {
			_ = r.Ws.WriteJSON(nostr.OKEnvelope{EventID: id, OK: false, Reason: "Error while fetching data from database"})
			fmt.Printf("Error while fetching data from database with: %v \n", err)
		}
		log.Println(events)

		if len(events) == 0 {
			//return fmt.Sprintf("Length of events is zero, no Events found for the given filter criteria")
		}
		for _, event := range events {
			_ = r.Ws.WriteJSON(nostr.EventEnvelope{SubscriptionID: &id, Event: event})
		}
	}

	_ = r.Ws.WriteJSON(nostr.EOSEEnvelope(id))
	// EOSE sent, now start streaming
	// TODO: need to add the subscription.
	// TODO: Should we store our subscriptions in database? mkaiser
	return ""

}
