package adapter

import (
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/handlers"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/models"
)

type LogAdapter struct {
}

/*
Create adapts the log to the model in the database and handles the insert.
*/
func (e *LogAdapter) Create(ip string, logType string, content string) (any, error) {
	logEntry := models.Log{IP: ip, Type: logType, Content: content}
	x, err := handlers.LogHandlerObject.CreateLogEntry(logEntry)
	return x, err
}
