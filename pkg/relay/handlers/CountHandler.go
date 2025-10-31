package handlers

import (
	"artio-relay/pkg/config"
	"artio-relay/pkg/storage/adapter"
	"artio-relay/pkg/webSocket"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

/*
CountHandler handles messages of type EVENT
*/
type CountHandler struct {
	Ctx any
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

/*
Handle handles the functionality where we request the count of events
*/
func (c CountHandler) Handle() string {
	// received count event with filters (handle similar to EventHandler
	var id string
	_ = json.Unmarshal(c.Req[1], &id)
	if id == "" {
		return "REQ has no <id>"
	}

	filters := make(nostr.Filters, len(c.Req)-2)
	for i, filterReq := range c.Req[2:] {
		filters[i].Limit = config.Config.RelayMaxMessageCount
		if err := json.Unmarshal(
			filterReq,
			&filters[i],
		); err != nil {
			return "failed to decode filter"
		}
	}

	IDs := make(map[string]bool)

	for _, filter := range filters {

		eventAdapter := adapter.EventAdapter{}
		var events []nostr.Event
		events, err := eventAdapter.Get(filter)

		if err != nil {
			_ = c.Ws.WriteJSON(nostr.OKEnvelope{EventID: id, OK: false, Reason: "Error while fetching data from database"})
			fmt.Printf("Error while fetching data from database with: %v \n", err)
			return ""
		}

		// go over all the events and set the value to true if found
		for _, event := range events {
			if !(IDs[event.ID]) {
				IDs[event.ID] = true
			}
		}
	}
	length := int64(len(IDs))
	_ = c.Ws.WriteJSON(nostr.CountEnvelope{Count: &length})
	_ = c.Ws.WriteJSON(nostr.EOSEEnvelope(id))
	return ""
}
