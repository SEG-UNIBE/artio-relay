package server

import (
	"encoding/json"
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
	"log"
	"net"
	"net/http"
	"nostr-relay/pkg/config"
	"nostr-relay/pkg/relay"
	"sync"
	"time"
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
}

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

	if err := s.httpServer.Serve(ln); err == http.ErrServerClosed {
		return nil
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

// ServeHTTP implements http.Handler interface.
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

func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("failed to upgrade websocket: %v", err)
		return
	}

	ip := conn.RemoteAddr().String()

	if realIP := r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP // possible to be multiple comma separated
	} else if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	log.Printf("connected from %s", ip)
	_, message, errMessage := conn.ReadMessage()
	if errMessage != nil {
		log.Fatalf("failed to read message: %v", errMessage)
	}
	var request []json.RawMessage
	_ = json.Unmarshal(message, &request)

	var typ string
	_ = json.Unmarshal(request[1], &typ)
	fmt.Println(typ)
	fmt.Println(string(request[1]))

}

func NewServer(relay *relay.Relay) *Server {
	// TODO: consider moving these to Server as config params
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	return &Server{relay: relay, upgrader: &upgrader}
}
