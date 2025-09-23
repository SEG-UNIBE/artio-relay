package handlers

import (
	"artio-relay/pkg/storage/models"
	"slices"

	"github.com/nbd-wtf/go-nostr"
)

type EventHandler struct {
	*BaseHandler
}

func (e EventHandler) CreateEvent(event models.Event) (any, error) {
	e.Connection.Table("events")
	results := e.Connection.Create(&event)
	return results, nil
}

/*
GetEvents fetches all the events out of the database and returns them as an array
*/
func (e EventHandler) GetEvents(filter nostr.Filter) ([]models.Event, error) {
	// we need to transfer all the needed values
	var results []models.Event
	e.Connection.Table("events")

	transaction := e.Connection

	if filter.IDs != nil {
		transaction = transaction.Where(map[string]interface{}{"event_id": filter.IDs})
	}

	if filter.Since != nil {
		transaction = transaction.Where("Created >= ?", filter.Since)
	}

	if filter.Until != nil {
		transaction = transaction.Where("Created <= ?", filter.Until)
	}

	if filter.Authors != nil {
		transaction = transaction.Where(map[string]interface{}{"pubkey": filter.Authors})
	}

	if filter.Kinds != nil {
		transaction = transaction.Where(map[string]interface{}{"kind": filter.Kinds})
	}

	if filter.Limit != 0 {
		transaction = transaction.Limit(filter.Limit)
	}

	// order the end result
	transaction.Order("Created desc")

	transaction.Find(&results)

	var outputResults []models.Event

	if len(filter.Tags) == 0 {
		// when there are no tag filters we return directly
		return results, nil
	}

	for i := range results {
		result := &results[i]

		toAppend := true

		for tagKey, tagValues := range filter.Tags {
			// loop over all the available tags
			tagFound := false
			// tagValues is a array
			for resultTagId := range result.Tags {
				tmpTag := result.Tags[resultTagId]
				if len(tmpTag) < 2 {
					// it is an invalid tag
					continue
				}
				if tmpTag[0] == tagKey && slices.Contains(tagValues, tmpTag[1]) {
					tagFound = true
					break
				}
			}
			toAppend = toAppend && tagFound
		}
		if toAppend {
			outputResults = append(outputResults, *result)
		}
	}
	return outputResults, nil
}

/*
DeleteEvent Handles the deletion of Events
*/
func (e EventHandler) DeleteEvent(event models.Event) error {
	e.Connection.Table("events")
	_ = e.Connection.Delete(event)
	return nil
}

/*
DeleteEvents Handles the deletion of Events
*/
func (e EventHandler) DeleteEvents(events []models.Event) error {
	e.Connection.Table("events")
	res := e.Connection.Delete(events)
	return res.Error
}

var baseHandler = NewBaseHandler([]models.Event{})
var EventHandlerObject = EventHandler{&baseHandler}
