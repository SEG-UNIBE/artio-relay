package adapter

import (
	"artio-relay/pkg/config"
	"artio-relay/pkg/storage/handlers"
	"artio-relay/pkg/storage/models"
	"github.com/nbd-wtf/go-nostr"
	"log"
	"strconv"
	"strings"
)

type EventAdapter struct {
}

/*
Create adapts the nostr event to the model in the database and handles the insert.
*/
func (e *EventAdapter) Create(event nostr.Event) (any, error) {
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

	if filter.Limit == 0 || filter.Limit > config.Config.RelayMaxMessageCount {
		// query only for the limited amount of events (order by time)
		filter.Limit = config.Config.RelayMaxMessageCount
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
Delete handles the deletion request
*/
func (e *EventAdapter) Delete(event nostr.Event) (error, bool) {
	deleteAllowed := true
	for _, tag := range event.Tags {
		// loop over all the tags
		var filter nostr.Filter
		if tag[0] == "e" {
			// delete by event id
			// only fetch the ones from database that we are actually allowed to delete
			filter = nostr.Filter{IDs: []string{tag[1]}}

		} else if tag[0] == "a" {
			values := tag[1]
			valueList := strings.Split(values, ":")

			if len(valueList) == 2 {
				kind, _ := strconv.ParseInt(valueList[0], 10, 32)
				pubkey := valueList[1]
				filter = nostr.Filter{Authors: []string{pubkey}, Kinds: []int{int(kind)}, Since: &event.CreatedAt}
			}

			if len(valueList) == 3 {
				kind, _ := strconv.ParseInt(valueList[0], 10, 32)
				pubkey := valueList[1]
				dIdentifier := valueList[2]
				tagMap := nostr.TagMap{"d": []string{dIdentifier}}
				filter = nostr.Filter{Authors: []string{pubkey}, Kinds: []int{int(kind)}, Tags: tagMap, Since: &event.CreatedAt}
			}
		}

		// take the predefined filter and start the fetching and deletion process
		result, err := handlers.EventHandlerObject.GetEvents(filter)

		if err != nil {
			log.Println(err)
			return err, false
		}

		for _, res := range result {
			if event.PubKey != res.Pubkey {
				deleteAllowed = false
				break
			}
		}

		if deleteAllowed {
			err = handlers.EventHandlerObject.DeleteEvents(result)
			if err != nil {
				log.Println(err)
				return err, false
			}
		}
	}
	return nil, deleteAllowed

}
