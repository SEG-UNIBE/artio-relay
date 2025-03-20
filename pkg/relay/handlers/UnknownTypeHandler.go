package handlers

import (
	"artio-relay/pkg/webSocket"
	"encoding/json"
)

/*
UnknownTypeHandler handles messages where the type is actually unknown
*/
type UnknownTypeHandler struct {
	Ctx any
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

/*
Handle handles the functionality for this event
*/
func (u UnknownTypeHandler) Handle() string {
	var typ string
	_ = json.Unmarshal(u.Req[0], &typ)
	return "the type " + typ + " is unknown"
}
