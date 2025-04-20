package handlers

import (
	"artio-relay/pkg/webSocket"
	"encoding/json"
	"github.com/nbd-wtf/go-nostr"
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

	var id string
	_ = json.Unmarshal(c.Req[1], &id)
	if id == "" {
		return "REQ has no <id>"
	}

	// TODO: remove listener
	_ = c.Ws.WriteJSON(nostr.NoticeEnvelope(""))
	return ""
}
