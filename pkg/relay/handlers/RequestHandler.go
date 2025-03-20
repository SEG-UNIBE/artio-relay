package handlers

import (
	"artio-relay/pkg/webSocket"
	"encoding/json"
	"errors"
	"log"
)

/*
RequestHandler handles messages of type REQ
*/
type RequestHandler struct {
	Ctx any
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

/*
Handle handles the functionality for this event
*/
func (r RequestHandler) Handle() string {
	err := errors.New("WRONG MESSAGE")
	log.Printf("error has occured %v", err)
	return "WRONG MESSAGE"
}
