package relay

import (
	"github.com/nbd-wtf/go-nostr/nip11"
	"reflect"
	"testing"
)

/*
TestRelayNIP11Local tests the return type of the NIP11 informations
*/
func TestRelayNIP11Local(t *testing.T) {
	r := Relay{}

	var nip11Informations = r.GetNIP11Information()

	if reflect.TypeOf(nip11Informations) != reflect.TypeOf(nip11.RelayInformationDocument{}) {
		t.Fatalf("the return nip11 information object does not have the correct type")
	}

}
