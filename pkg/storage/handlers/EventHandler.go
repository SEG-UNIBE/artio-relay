package handlers

import (
	"artio-relay/pkg/storage/models"
)

type EventHandler struct {
	*BaseHandler
}

func (e EventHandler) CreateEvent(event models.Event) (any, error) {
	e.Connection.Table("events") // TODO: Fix this stupid thing that will no work
	results := e.Connection.Create(&event)
	return results, nil
}

var baseHandler = NewBaseHandler([]models.Event{})
var EventHandlerObject = EventHandler{&baseHandler}
