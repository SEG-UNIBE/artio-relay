package handlers

import (
	"artio-relay/pkg/storage/models"
	"github.com/nbd-wtf/go-nostr"
	"slices"
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

	// TODO: need to transform this into a usable event return type
	transaction := e.Connection

	if filter.IDs != nil {
		transaction = transaction.Where(map[string]interface{}{"id": filter.IDs})
	}

	if filter.Since != nil {
		transaction = transaction.Where("Created >= ?", filter.Since)
	}

	if filter.Until != nil {
		transaction = transaction.Where("Created < ?", filter.Until)
	}

	if filter.Authors != nil {
		transaction = transaction.Where(map[string]interface{}{"pubkey": filter.Authors})
	}

	if filter.Kinds != nil {
		transaction = transaction.Where(map[string]interface{}{"kind": filter.Kinds})
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
		appended := false

		for tagKey, tagValues := range filter.Tags {
			// loop over all the available tags
			// tagValues is a array
			for resultTagId := range result.Tags {
				tmpTag := result.Tags[resultTagId]
				if len(tmpTag) < 2 {
					// it is an invalid tag
					continue
				}
				if tmpTag[0] == tagKey && slices.Contains(tagValues, tmpTag[1]) {
					outputResults = append(outputResults, *result)
					appended = true
					break
				}
			}
			if appended {
				break
			}
		}
	}
	return outputResults, nil
}

var baseHandler = NewBaseHandler([]models.Event{})
var EventHandlerObject = EventHandler{&baseHandler}
