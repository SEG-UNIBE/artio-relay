package server

import (
	"artio-relay/pkg/config"
	"artio-relay/pkg/relay"
	"artio-relay/pkg/webSocket"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

/*
challenge creating a websocket with a cryptographic challenge
*/
func challenge(conn *websocket.Conn) *webSocket.WebSocket {
	// NIP-42 challenge
	challenge := make([]byte, 8)
	rand.Read(challenge)

	return &webSocket.WebSocket{
		Conn:      conn,
		Challenge: hex.EncodeToString(challenge),
	}
}

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
}

/*
Start is the function to startup the server and handle then delegate the handling of the traffic
*/
func (s *Server) Start() error {
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

	} else if r.Header.Get("Upgrade") == "websocket" {
		s.HandleWebsocket(w, r)
	} else {
		log.Fatal("Not implemented") // TODO: implement this
	}

	/*
		if r.Header.Get("Upgrade") == "websocket" {
			s.HandleWebsocket(w, r)
		} else if r.Header.Get("Accept") == "application/nostr+json" {
			s.HandleNIP11(w, r)
		} else {
			s.serveMux.ServeHTTP(w, r)
		}
	*/
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
	log.Printf("connected from %s", ip)
	_, message, errMessage := conn.ReadMessage()
	if errMessage != nil {
		log.Printf("failed to read message: %v", errMessage)
	}
	var request []json.RawMessage
	_ = json.Unmarshal(message, &request)

	ws := challenge(conn)
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
				//removeListener(ws)
			}
			s.clientsMu.Unlock()
			log.Printf("disconnected from %s\n", ip)
		}()

		// set some limits on the connection to assure the correct functionality
		conn.SetReadLimit(config.Config.RelayMaxMessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(config.Config.RelayPongWait))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(config.Config.RelayPongWait))
			return nil
		})
		defer cancel()

		// NIP-42 auth challenge
		//if _, ok := s.relay.(Auther); ok {
		//	ws.WriteJSON(nostr.AuthEnvelope{Challenge: &ws.challenge})
		//}

		for {
			typ, message, err := conn.ReadMessage()
			fmt.Println("typ:", typ, "message:", string(message), "err:", err)
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

			go s.relay.HandleMessage(ctx, ws, message)
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
				log.Printf("pinging for %s", ip)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func NewServer(relay *relay.Relay) *Server {
	return &Server{relay: relay, upgrader: webSocket.NewUpgrader()}
}
