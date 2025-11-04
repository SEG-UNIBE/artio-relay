package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/logging"
	"github.com/SEG-UNIBE/artio-relay/pkg/relay"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/adapter"
	"github.com/SEG-UNIBE/artio-relay/pkg/webSocket"
	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
)

/*
IServer to specify the Server functionalities
*/
type IServer interface {
	Start() error
}

/*
Server is the base for the complete implementation of our nostr relay
It will handle the incoming requests and delegate it to the corresponding functionalities in the backend e.g. relay
*/
type Server struct {
	IServer
	relay *relay.Relay

	upgrader *websocket.Upgrader

	// in case you call Server.Start
	Addr       string
	serveMux   *http.ServeMux
	httpServer *http.Server

	clientsMu sync.Mutex
	clients   map[*websocket.Conn]struct{}

	LogAdapter adapter.LogAdapter
}

/*
Start is the function to start up the server and handle then delegate the handling of the traffic
*/
func (s *Server) Start() error {
	// create the serveMux if not already done
	if s.serveMux == nil {
		s.serveMux = http.NewServeMux()
	}

	addr := config.Config.GetRelayAddress()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.Addr = ln.Addr().String()
	s.httpServer = &http.Server{
		Handler:      cors.Default().Handler(s),
		Addr:         addr,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := s.httpServer.Serve(ln); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

/*
ServeHTTP implements http.Handler interface.
*/
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Accept") == "application/nostr+json" {
		w.Header().Set("Content-Type", "application/json")
		info := s.relay.GetNIP11Information()

		_ = json.NewEncoder(w).Encode(info)
		go logging.ArtioLogger.LogNIP11(r.RemoteAddr)

	} else if r.Header.Get("Upgrade") == "websocket" {
		s.HandleWebsocket(w, r)
	} else {
		s.serveMux.ServeHTTP(w, r)
	}
}

/*
HandleWebsocket function to exract and delegate traffic from websockets.
*/
func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("failed to upgrade websocket: %v", err)
		return
	}

	//s.clientsMu.Lock()
	//defer s.clientsMu.Unlock()
	//s.clients[conn] = struct{}{}
	ticker := time.NewTicker(config.Config.RelayPingWait)

	ip := conn.RemoteAddr().String()

	if realIP := r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP // possible to be multiple comma separated
	} else if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	go logging.ArtioLogger.LogConnect(ip)

	ws := s.relay.Challenge(conn)
	ctx, cancel := context.WithCancel(context.Background())

	// reader
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			s.clientsMu.Lock()
			if _, ok := s.clients[conn]; ok {
				_ = conn.Close()
				delete(s.clients, conn)
			}
			s.clientsMu.Unlock()
			go logging.ArtioLogger.LogDisconnect(ip)
		}()

		// set some limits on the connection to assure the correct functionality
		conn.SetReadLimit(config.Config.RelayMaxMessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(config.Config.RelayPongWait))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(config.Config.RelayPongWait))
			return nil
		})
		defer cancel()

		s.relay.SendAuthMessage(ws)

		for {
			typ, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("typ:", typ, "message:", string(message), "err:", err)
			} else {
				go logging.ArtioLogger.LogRequest("GENREQ", string(message), ip)
			}

			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,        // 1001
					websocket.CloseNoStatusReceived, // 1005
					websocket.CloseAbnormalClosure,  // 1006
				) {
					//s.Log.Warningf("unexpected close error from %s: %v", r.Header.Get("X-Forwarded-For"), err)
				}
				break
			}

			if typ == websocket.PingMessage {
				_ = ws.WriteMessage(websocket.PongMessage, nil)
				continue
			}

			go s.relay.HandleMessage(&ctx, ws, message)
		}
	}()

	// writer
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			_ = conn.Close()
		}()

		for {
			select {
			case <-ticker.C:
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(config.Config.RelayWriteWait))
				if err != nil {
					log.Printf("error writing ping: %v; closing websocket", err)
					return
				}
				go logging.ArtioLogger.LogPing(ip)
			case <-ctx.Done():
				return
			}
		}
	}()
}

/*
InjectHandler allows to inject custom http.Handlers into the server on a given path
this will panic if the pattern is already registered
*/
func (s *Server) InjectHandler(pattern string, handler http.Handler) {
	s.serveMux.Handle(pattern, handler)
}

func NewServer(relay *relay.Relay) *Server {
	return &Server{
		relay:      relay,
		upgrader:   webSocket.NewUpgrader(),
		LogAdapter: adapter.LogAdapter{},
		serveMux:   http.NewServeMux()}
}
