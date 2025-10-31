package server

import (
	"testing"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/relay"
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
