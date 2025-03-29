package handlers

import (
	"artio-relay/pkg/storage/models"
	"github.com/nbd-wtf/go-nostr"
)

type EventHandler struct {
	*BaseHandler
}

func (e EventHandler) CreateEvent(event models.Event) (any, error) {
	e.Connection.Table("events") // TODO: Fix this stupid thing that will no work
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
	e.Connection.Where(map[string]interface{}{"id": filter.IDs}).Where(map[string]interface{}{"pubkey": filter.Authors}).Where(map[string]interface{}{"kind": filter.Kinds}).Find(&results)

	return results, nil
}

var baseHandler = NewBaseHandler([]models.Event{})
var EventHandlerObject = EventHandler{&baseHandler}
