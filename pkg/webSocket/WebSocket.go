package webSocket

import (
	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
	"sync"
)

type WebSocket struct {
	Conn  *websocket.Conn
	mutex sync.Mutex

	// nip42
	Challenge string
	authed    string
	limiter   *rate.Limiter
}

func (ws *WebSocket) WriteJSON(any interface{}) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteJSON(any)
}

func (ws *WebSocket) WriteMessage(t int, b []byte) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteMessage(t, b)
}
