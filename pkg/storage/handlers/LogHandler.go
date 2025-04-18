package handlers

import (
	"artio-relay/pkg/storage/models"
)

type LogHandler struct {
	*BaseHandler
}

/*
CreateLogEntry creates an entry in the database for a log
*/
func (e EventHandler) CreateLogEntry(logEntry models.Log) (any, error) {
	e.Connection.Table("logs")
	results := e.Connection.Create(&logEntry)
	return results, nil
}

var baseHandlerLog = NewBaseHandler([]models.Event{})
var LogHandlerObject = EventHandler{&baseHandlerLog}
