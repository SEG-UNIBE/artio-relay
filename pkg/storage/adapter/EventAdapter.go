package adapter

import (
	"artio-relay/pkg/storage/handlers"
	"artio-relay/pkg/storage/models"
	"github.com/nbd-wtf/go-nostr"
)

type EventAdapter struct {
}

/*
Create adapts the nostr event to the model in the database and handles the insert.
*/
func (e *EventAdapter) Create(event nostr.Event) (any, error) {
	eventModel := models.Event{Id: event.ID, Pubkey: event.PubKey, Kind: uint32(event.Kind), Sig: event.Sig, Content: event.Content}
	x, err := handlers.EventHandlerObject.CreateEvent(eventModel)
	return x, err
}
