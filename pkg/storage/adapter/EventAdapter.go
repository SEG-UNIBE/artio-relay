package adapter

import (
	"log"
	"strconv"
	"strings"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/handlers"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/models"

	"github.com/nbd-wtf/go-nostr"
)

type EventAdapter struct {
}

/*
Create adapts the nostr event to the model in the database and handles the insert.
*/
func (e *EventAdapter) Create(event nostr.Event) (any, error) {
	if event.Tags == nil {
		// the json serializer turns nil into "" for NOT NULL columns, which is invalid jsonb
		event.Tags = nostr.Tags{}
	}
	eventModel := models.Event{
		EventId: event.ID,
		Created: event.CreatedAt.Time().Unix(),
		Pubkey:  event.PubKey,
		Kind:    uint32(event.Kind),
		Sig:     event.Sig,
		Content: event.Content,
		Tags:    event.Tags,
	}
	x, err := handlers.EventHandlerObject.CreateEvent(eventModel)
	return x, err
}

/*
Get all the events out of the database for a given nostr.Filter
*/
func (e *EventAdapter) Get(filter nostr.Filter) ([]nostr.Event, error) {
	// TODO implement the adapter functionality
	// should translate nostr.filter into a gorm understandable model

	if filter.Limit > config.Config.RelayMaxMessageCount {
		// query only for the limited amount of events (order by time)
		filter.Limit = config.Config.RelayMaxMessageCount
	}

	var events []nostr.Event
	// fetching the intermediate (ir) results from the database
	var irResults, err = handlers.EventHandlerObject.GetEvents(filter)

	if err != nil {
		return nil, err
	}

	for _, result := range irResults {
		// handling the max amount of results to return
		if len(events) >= filter.Limit {
			return events, nil
		}
		tmpEvent := nostr.Event{
			ID:        result.EventId,
			PubKey:    result.Pubkey,
			CreatedAt: nostr.Timestamp(result.Created),
			Kind:      int(result.Kind),
			Content:   result.Content,
			Sig:       result.Sig,
			Tags:      result.Tags,
		}
		events = append(events, tmpEvent)

	}
	if irResults == nil {
		log.Println("No results found")
	}

	return events, nil
}

/*
Delete handles the deletion request. Every tag is applied independently, the returned bool is false
when at least one tag referenced events of another author, which were skipped.
*/
func (e *EventAdapter) Delete(event nostr.Event) (error, bool) {
	allPermitted := true
	for _, tag := range event.Tags {
		// loop over all the tags, skipping malformed ones
		if len(tag) < 2 {
			continue
		}
		var filter nostr.Filter
		if tag[0] == "e" {
			// delete by event id
			// only fetch the ones from database that we are actually allowed to delete
			filter = nostr.Filter{Authors: []string{event.PubKey}, IDs: []string{tag[1]}}
		} else if tag[0] == "a" {
			// delete by address "<kind>:<pubkey>:<d-identifier>" (the legacy "<kind>:<pubkey>" form is also accepted)
			valueList := strings.Split(tag[1], ":")
			if len(valueList) != 2 && len(valueList) != 3 {
				// malformed address, nothing to delete
				continue
			}

			kind, err := strconv.Atoi(valueList[0])
			if err != nil {
				// malformed kind, nothing to delete
				continue
			}

			if valueList[1] != event.PubKey {
				// only the author of an address may delete it, skip it but keep applying the other tags
				allPermitted = false
				continue
			}

			// NIP-09: delete all versions up to the created_at of the deletion request
			filter = nostr.Filter{Authors: []string{event.PubKey}, Kinds: []int{kind}, Until: &event.CreatedAt}

			if kind >= 30000 && kind < 40000 {
				// addressable events are identified by their d tag, an empty identifier is a valid value
				dIdentifier := ""
				if len(valueList) == 3 {
					dIdentifier = valueList[2]
				}
				filter.Tags = nostr.TagMap{"d": []string{dIdentifier}}
			}
		} else {
			continue
		}
		// take the predefined filter and start the fetching and deletion process
		// the filters only match events of the requesting author, so everything found may be deleted
		result, err := handlers.EventHandlerObject.GetEvents(filter)

		if err != nil {
			log.Println(err)
			return err, false
		}

		err = handlers.EventHandlerObject.DeleteEvents(result)
		if err != nil {
			log.Println(err)
			return err, false
		}
	}
	return nil, allPermitted

}

/*
DeleteAndInsertKind3 handles all Kind 3 events for NIP-02. delete all follow lists for the given pubkey and create new one.
*/
func (e *EventAdapter) DeleteAndInsertKind3(event nostr.Event) error {
	pubkey := event.PubKey
	filter := nostr.Filter{Kinds: []int{3}, Authors: []string{pubkey}}
	var irResults, err = handlers.EventHandlerObject.GetEvents(filter)
	if err != nil {
		return err
	}
	err = handlers.EventHandlerObject.DeleteEvents(irResults)
	if err != nil {
		return err
	}
	_, err = e.Create(event)
	return err
}
