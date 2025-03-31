package webSocket

import (
	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
	"sync"
)

/*
WebSocket type definition.
*/
type WebSocket struct {
	Conn  *websocket.Conn
	mutex sync.Mutex

	// nip42
	Challenge string
	authed    string
	limiter   *rate.Limiter
}

/*
WriteJSON writes a output JSON to the websocket with mutual exclusion activated.
*/
func (ws *WebSocket) WriteJSON(any interface{}) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteJSON(any)
}

/*
WriteMessage writes a output to the websocket with mutual exclusion activated.
*/
func (ws *WebSocket) WriteMessage(t int, b []byte) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteMessage(t, b)
}
