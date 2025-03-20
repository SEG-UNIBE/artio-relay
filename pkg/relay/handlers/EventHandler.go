package handlers

import (
	"artio-relay/pkg/webSocket"
	"encoding/json"
	"errors"
	"log"
)

/*
EventHandler handles messages of type EVENT
*/
type EventHandler struct {
	Ctx any
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

/*
Handle handles the functionality for this event
*/
func (e EventHandler) Handle() string {
	err := errors.New("WRONG MESSAGE")
	log.Fatalf("error has occured %v", err)
	return ""
}
