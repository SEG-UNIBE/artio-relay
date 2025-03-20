package handlers

import (
	"artio-relay/pkg/storage/models"
)

type EventHandler struct {
	*BaseHandler
}

var baseHandler = NewBaseHandler([]models.Event{})
var EventHandlerObject = EventHandler{&baseHandler}
