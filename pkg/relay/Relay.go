package relay

import (
	"encoding/json"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"
	"log"
	"nostr-relay/pkg/config"
	"nostr-relay/pkg/storage"
	"nostr-relay/pkg/webSocket"
)

type IRelay interface {
	GetNIP11Information()
	HandleEvent()
	HandleMessage()
}

type Relay struct {
	Storage storage.Storage
	Name    string
}

func (relay *Relay) GetNIP11Information() nip11.RelayInformationDocument {

	// supportedNIPs := []any{9, 11, 12, 15, 16, 20, 33}
	supportedNIPs := []any{11}
	// TODO: Implement the NIP42
	/*
		if _, ok := s.relay.(Auther); ok {
			supportedNIPs = append(supportedNIPs, 42)
		}
	*/
	// TODO: Implement the NIP45
	/*
		if storage, ok := s.relay.(eventstore.Store); ok && storage != nil {
			if _, ok = storage.(EventCounter); ok {
				supportedNIPs = append(supportedNIPs, 45)
			}
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
	}
}

func (relay *Relay) HandleMessage(ctx any, ws *webSocket.WebSocket, message []byte) {
	var notice string
	// function gets executed after the rest of the function is done.
	defer func() {
		if notice != "" {
			err := ws.WriteJSON(nostr.NoticeEnvelope(notice))
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
	fmt.Println(typ)
	fmt.Println(request)
	//switch typ {
	//case "EVENT":
	//	notice = relay.doEvent(ctx, ws, request, relay.Storage)
	//case "REQ":
	//	notice = relay.doReq(ctx, ws, request, relay.Storage)
	//case "CLOSE":
	//	notice = relay.doClose(ctx, ws, request, relay.Storage)
	//default:
	//	if cwh, ok := relay.(CustomWebSocketHandler); ok {
	//		cwh.HandleUnknownType(ws, typ, request)
	//	} else {
	//		notice = "unknown message type " + typ
	//	}
	//}
}

/*
NewRelay creates a new relay together with the provided storage
*/
func NewRelay(storage storage.Storage) *Relay {
	return &Relay{storage, config.Config.RelayName()}
}
