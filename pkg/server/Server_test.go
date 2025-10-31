package server

import (
	"artio-relay/pkg/config"
	"artio-relay/pkg/relay"
	"testing"
)

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
