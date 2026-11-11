package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SEG-UNIBE/artio-relay/pkg/storage/models"

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
		transaction = transaction.Where("created >= ?", filter.Since)
	}

	if filter.Until != nil {
		transaction = transaction.Where("created <= ?", filter.Until)
	}

	if filter.Authors != nil {
		transaction = transaction.Where(map[string]interface{}{"pubkey": filter.Authors})
	}

	if filter.Kinds != nil {
		transaction = transaction.Where(map[string]interface{}{"kind": filter.Kinds})
	}

	if filter.Search != "" {
		transaction = transaction.Where("content LIKE ?", fmt.Sprintf("%%%s%%", filter.Search))
	}

	// tag filters are applied in the database (before the limit), each key must match one of its values
	for tagKey, tagValues := range filter.Tags {
		if len(tagValues) == 0 {
			// a tag filter without values can never match
			return nil, nil
		}

		// containment check that can use the GIN index on tags
		containment := make([]string, 0, len(tagValues))
		args := make([]any, 0, len(tagValues))
		for _, tagValue := range tagValues {
			needle, err := json.Marshal([][]string{{tagKey, tagValue}})
			if err != nil {
				return nil, err
			}
			containment = append(containment, "tags @> ?::jsonb")
			args = append(args, string(needle))
		}
		transaction = transaction.Where("("+strings.Join(containment, " OR ")+")", args...)

		// jsonb containment ignores element positions, so verify key and value are the first two entries of a tag
		transaction = transaction.Where(
			"EXISTS (SELECT 1 FROM jsonb_array_elements(tags) AS tag WHERE tag->>0 = ? AND tag->>1 IN ?)",
			tagKey, []string(tagValues),
		)
	}

	// newest first, ties broken by lowest id as specified in NIP-01
	transaction = transaction.Order("created desc").Order("event_id asc")

	// gorm emits LIMIT 0 for a zero limit, so only limit when one was set or explicitly requested as 0
	if filter.Limit > 0 || filter.LimitZero {
		transaction = transaction.Limit(filter.Limit)
	}

	if err := transaction.Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
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
	if len(events) == 0 {
		// gorm refuses a delete without conditions, and there is nothing to delete anyway
		return nil
	}
	res := e.Connection.Delete(events)
	return res.Error
}

var baseHandler = NewBaseHandler([]models.Event{})
var EventHandlerObject = EventHandler{&baseHandler}
