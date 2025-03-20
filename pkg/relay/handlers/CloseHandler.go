package handlers

import (
	"artio-relay/pkg/webSocket"
	"encoding/json"
	"errors"
	"log"
)

/*
CloseHandler handles messages of type CLOSE
*/
type CloseHandler struct {
	Ctx any
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

/*
Handle handles the functionality for this event
*/
func (c CloseHandler) Handle() string {
	err := errors.New("WRONG MESSAGE")
	log.Fatalf("error has occured %v", err)
	return ""
}
