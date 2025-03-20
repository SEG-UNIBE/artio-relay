package server

import (
	"nostr-relay/pkg/config"
	"nostr-relay/pkg/relay"
	"reflect"
	"testing"
)

/*
TestServerChallenge tests if the return value of the server challenge is of valid type
*/
func TestServerChallenge(t *testing.T) {
	socket := challenge(nil)

	if socket.Conn != nil {
		t.Fatalf("the socket connection is not nil: %v", socket.Conn)
	}

	if reflect.TypeOf(socket.Challenge).Kind() != reflect.String {
		t.Fatalf("challenge is not a string: %v", reflect.TypeOf(socket.Challenge))
	}
}

func funcTestServerSetup(t *testing.T) {
	server := NewServer(&relay.Relay{})
	err := server.Start()

	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	if server.Addr != config.Config.GetRelayAddress() {
		t.Fatalf("the relay address is not correct: %v", server.Addr)
	}
}
