package webSocket

import (
	"context"
	"sync"
	"time"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/stats"
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
NewWebSocket creates a WebSocket with a per-connection rate limiter.
*/
func NewWebSocket(conn *websocket.Conn, challenge string, serviceURL string) *WebSocket {
	ws := &WebSocket{
		Conn:       conn,
		Challenge:  challenge,
		ServiceURL: serviceURL,
	}
	if config.Config.RelayMessageRate > 0 && config.Config.RelayMessageBurst > 0 {
		ws.limiter = rate.NewLimiter(rate.Limit(config.Config.RelayMessageRate), config.Config.RelayMessageBurst)
	}
	return ws
}

/*
setWriteDeadline bounds how long a single write may block on a slow client.
*/
func (ws *WebSocket) setWriteDeadline() {
	if config.Config.RelayWriteWait > 0 {
		_ = ws.Conn.SetWriteDeadline(time.Now().Add(config.Config.RelayWriteWait))
	}
}

/*
WaitForSlot blocks until the rate limiter allows another message or the context is cancelled.
*/
func (ws *WebSocket) WaitForSlot(ctx context.Context) error {
	if ws.limiter == nil {
		return nil
	}
	return ws.limiter.Wait(ctx)
}

/*
WriteJSON writes an output JSON to the websocket with mutual exclusion activated.
*/
func (ws *WebSocket) WriteJSON(any interface{}) error {
	go stats.MessageOut()
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.setWriteDeadline()
	return ws.Conn.WriteJSON(any)
}

/*
WriteMessage writes an output to the websocket with mutual exclusion activated.
*/
func (ws *WebSocket) WriteMessage(t int, b []byte) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.setWriteDeadline()
	return ws.Conn.WriteMessage(t, b)
}

/*
GetRemoteIP returns the IP Address of the remote end of the connection
*/
func (ws *WebSocket) GetRemoteIP() string {
	return ws.Conn.RemoteAddr().String()
}
