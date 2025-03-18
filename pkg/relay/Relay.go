package relay

import (
	"github.com/nbd-wtf/go-nostr/nip11"
	"nostr-relay/pkg/config"
	"nostr-relay/pkg/storage"
)

type IRelay interface {
	GetNIP11Information()
	HandleEvent()
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

/*
NewRelay creates a new relay together with the provided storage
*/
func NewRelay(storage storage.Storage) *Relay {
	return &Relay{storage, config.Config.RelayName()}
}
