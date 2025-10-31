package relay

import (
	"reflect"
	"testing"

	"github.com/nbd-wtf/go-nostr/nip11"
)

/*
TestRelayNIP11Local tests the return type of the NIP11 information
*/
func TestRelayNIP11Local(t *testing.T) {
	r := Relay{}

	var nip11Informations = r.GetNIP11Information()

	if reflect.TypeOf(nip11Informations) != reflect.TypeOf(nip11.RelayInformationDocument{}) {
		t.Fatalf("the return nip11 information object does not have the correct type")
	}
}

/*
TestRelayChallenge tests if the return value of the relay challenge is of valid type
*/
func TestRelayChallenge(t *testing.T) {
	relay := Relay{}
	socket := relay.Challenge(nil)

	if socket.Conn != nil {
		t.Fatalf("the socket connection is not nil: %v", socket.Conn)
	}

	if reflect.TypeOf(socket.Challenge).Kind() != reflect.String {
		t.Fatalf("challenge is not a string: %v", reflect.TypeOf(socket.Challenge))
	}
}
