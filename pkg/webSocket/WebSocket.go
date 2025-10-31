package webSocket

import (
	"sync"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
)

/*
WebSocket type definition.
*/
type WebSocket struct {
	Conn  *websocket.Conn
	mutex sync.Mutex

	// nip42
	Challenge     string
	Authenticated string
	ServiceURL    string
	limiter       *rate.Limiter
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

/*
GetRemoteIP returns the IP Address of the remote end of the connection
*/
func (ws *WebSocket) GetRemoteIP() string {
	return ws.Conn.RemoteAddr().String()
}
