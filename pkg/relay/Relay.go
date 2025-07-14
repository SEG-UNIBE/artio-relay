package relay

import (
	"artio-relay/pkg/config"
	"artio-relay/pkg/logging"
	"artio-relay/pkg/relay/handlers"
	"artio-relay/pkg/storage"
	"artio-relay/pkg/webSocket"
	"context"
	"encoding/json"
	"github.com/nbd-wtf/go-nostr/nip11"
	"log"
)

type IRelay interface {
	GetNIP11Information() nip11.RelayInformationDocument
	HandleMessage(ctx any, ws *webSocket.WebSocket, message []byte)
}

type Relay struct {
	Storage storage.Storage
	Name    string
}

func (relay *Relay) GetNIP11Information() nip11.RelayInformationDocument {

	// supportedNIPs := []any{9, 11, 12, 15, 16, 20, 33}
	supportedNIPs := []any{9, 11, 45, 65}
	// TODO: Implement the NIP42
	/*
		if _, ok := s.relay.(Auther); ok {
			supportedNIPs = append(supportedNIPs, 42)
		}
	*/

	return nip11.RelayInformationDocument{
		Name:          relay.Name,
		Description:   config.Config.NIP11Description,
		PubKey:        config.Config.NIP11Pubkey,
		Contact:       config.Config.NIP11Contact,
		SupportedNIPs: supportedNIPs,
		Software:      config.Config.NIP11Software,
		Version:       config.Config.NIP11Version,
		Icon:          config.Config.NIP11Banner,
		Banner:        config.Config.NIP11Banner,
	}
}

func (relay *Relay) HandleMessage(ctx context.Context, ws *webSocket.WebSocket, message []byte) {
	var notice string
	// function gets executed after the rest of the function is done.
	defer func() {
		if notice != "" {
			err := ws.WriteJSON(notice)
			if err != nil {
				log.Fatalf("error writing JSON: %v", err)
			}
		}
	}()

	var request []json.RawMessage
	if err := json.Unmarshal(message, &request); err != nil {
		// stop silently
		return
	}

	if len(request) < 2 {
		notice = "request has less than 2 parameters"
		return
	}

	var typ string
	_ = json.Unmarshal(request[0], &typ)

	var handler handlers.Handler
	logging.ArtioLogger.LogHandling("RELAYHANDLE", typ, ws.GetRemoteIP())
	switch typ {
	case "EVENT":
		handler = handlers.EventHandler{Ctx: ctx, Ws: ws, Req: request}
	case "COUNT":
		handler = handlers.CountHandler{Ctx: ctx, Ws: ws, Req: request}
	case "REQ":
		handler = handlers.RequestHandler{Ctx: ctx, Ws: ws, Req: request}
	case "CLOSE":
		handler = handlers.CloseHandler{Ctx: ctx, Ws: ws, Req: request}
	default:
		handler = handlers.UnknownTypeHandler{Ctx: ctx, Ws: ws, Req: request}
	}
	notice = handler.Handle()
}

/*
NewRelay creates a new relay together with the provided storage
*/
func NewRelay(storage storage.Storage) *Relay {
	return &Relay{storage, config.Config.RelayName()}
}
