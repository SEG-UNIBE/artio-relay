package adapter

import (
	"artio-relay/pkg/storage/handlers"
	"artio-relay/pkg/storage/models"
	"github.com/nbd-wtf/go-nostr"
	"log"
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

/*
Get all the events out of the database for a given nostr.Filter
*/
func (e *EventAdapter) Get(filter nostr.Filter) ([]nostr.Event, error) {
	// TODO implement the adapter functionality
	// should translate nostr.filter into a gorm understandable model

	if filter.Limit == 0 {
		// query only for the limited amount of events (order by time)
	}

	var events []nostr.Event
	// fetching the intermediate (ir) results from the database
	var irResults, err = handlers.EventHandlerObject.GetEvents(filter)

	if err != nil {
		return nil, err
	}

	if filter.Limit == 0 {
		filter.Limit = 999999
	}

	for _, result := range irResults {

		// handling the max amount of results to return
		if len(events) >= filter.Limit {
			return events, nil
		}

		// type Event struct {
		// 	ID        string
		// 	PubKey    string
		// 	CreatedAt Timestamp
		// 	Kind      int
		// 	Tags      Tags
		// 	Content   string
		// 	Sig       string
		// }
		tmpEvent := nostr.Event{
			ID:        result.Id,
			PubKey:    result.Pubkey,
			CreatedAt: nostr.Timestamp(result.Created),
			Kind:      int(result.Kind),
			Content:   result.Content,
			Sig:       result.Sig,
		}
		events = append(events, tmpEvent)

	}
	if irResults == nil {
		log.Println("No results found")
	}

	return events, nil
}
