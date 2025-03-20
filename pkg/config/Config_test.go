package config

import (
	"testing"
	"time"
)

/*
TestConfigDatabaseParams tests the database parameters of the config
*/
func TestConfigDatabaseParams(t *testing.T) {
	var config, err = (&(Configuration{})).Init("../../.env.tests")

	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	if config.postgresHost != "127.0.0.1" {
		t.Fatalf("postgres host should be 127.0.0.1 but got %v", config.postgresHost)
	}

	if config.postgresUser != "nostr" {
		t.Fatalf("postgres user should be nostr but got %v", config.postgresUser)
	}

	if config.postgresPassword != "nostr" {
		t.Fatalf("postgres password should be nostr but got %v", config.postgresPassword)
	}

	if config.postgresDatabase != "nostr" {
		t.Fatalf("postgres database should be nostr but got %v", config.postgresDatabase)
	}

	if config.postgresPort != "5432" {
		t.Fatalf("postgres port should be 5432 but got %v", config.postgresPort)
	}
}

/*
TestConfigRelayParams tests the relay parameters of the config
*/
func TestConfigRelayParams(t *testing.T) {
	var config, err = (&(Configuration{})).Init("../../.env.tests")

	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	if config.relayAddress != "0.0.0.0" {
		t.Fatalf("relay address should be 0.0.0.0 but got %v", config.postgresHost)
	}

	if config.relayPort != "8000" {
		t.Fatalf("relay port should be 8000but got %v", config.postgresUser)
	}

}

/*
TestConfigRNIP11Params tests the nip11 parameters of the config
*/
func TestConfigRNIP11Params(t *testing.T) {
	var config, err = (&(Configuration{})).Init("../../.env.tests")

	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	if config.NIP11Version != "0.9" {
		t.Fatalf("NIP11 version should be 0.9 but got %v", config.NIP11Version)
	}

	if config.NIP11Software != "https://github.com/SEG-UNIBE/artio-relay" {
		t.Fatalf("NIP11 software url does not match but got %v", config.NIP11Software)
	}

	if config.NIP11Description != "artio relayer implementation for the university of bern" {
		t.Fatalf("NIP11 description does not match got %v", config.NIP11Description)
	}

	if config.NIP11Pubkey != "~~" {
		t.Fatalf("NIP11 Pubkey does not match got %v", config.NIP11Pubkey)
	}

	if config.NIP11Contact != "~" {
		t.Fatalf("NIP11 Contact does not match got %v", config.NIP11Contact)
	}

}

/*
TestGetRelayAddress tests the relay address constructor
*/
func TestGetRelayAddress(t *testing.T) {
	var config, err = (&(Configuration{})).Init("../../.env.tests")

	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	var connString = config.GetRelayAddress()

	if connString != "0.0.0.0:8000" {
		t.Fatalf("0.0.0.0:8000 but got %v", connString)
	}
}

/*
TestGetRelayParams tests the server parameters of the config
*/
func TestGetRelayParams(t *testing.T) {
	var config, err = (&(Configuration{})).Init("../../.env.tests")

	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	if config.RelayWriteWait != 10*time.Second {
		t.Fatalf("RelayWriteWait does not have the correct value, got: %v", config.RelayWriteWait)
	}

	if config.RelayPongWait != 60*time.Second {
		t.Fatalf("RelayPongWait does not have the correct value, got: %v", config.RelayPongWait)
	}

	if config.RelayPingWait != 30*time.Second {
		t.Fatalf("RelayPingWait does not have the correct value, got: %v", config.RelayPingWait)
	}

	if config.RelayPingWait != config.RelayPongWait/2 {
		t.Fatalf("RelayPingWait is not equal to half of RelayPongWait, got: %v", config.RelayPingWait)
	}
}

/*
TestConfigGetConnectionString tests the database connection string constructor
*/
func TestConfigGetConnectionString(t *testing.T) {
	var config, err = (&(Configuration{})).Init("../../.env.tests")

	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	var connString = config.GetDatabaseConnectionString()

	if connString != "host=127.0.0.1 user=nostr password=nostr dbname=nostr port=5432 sslmode=disable TimeZone=UTC" {
		t.Fatalf("postgres host should be 127.0.0.1 but got %v", connString)
	}
}
