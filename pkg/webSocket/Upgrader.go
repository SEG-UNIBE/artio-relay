package webSocket

import (
	"net/http"

	"github.com/fasthttp/websocket"
)

/*
NewUpgrader generates a websocket Upgrader that is used in the server
*/
func NewUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
}
