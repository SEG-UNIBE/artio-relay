package config

import (
	"testing"
)

// note: godotenv never overrides variables that are already set, so t.Setenv takes precedence over .env.tests

/*
TestConfigMissingEnvFile tests that a missing env file is reported as an error
*/
func TestConfigMissingEnvFile(t *testing.T) {
	_, err := (&Configuration{}).Init("../../does-not-exist.env")

	if err == nil {
		t.Fatalf("expected an error for a missing env file")
	}
}

/*
TestConfigMaxMessageSizeFromEnv tests that the max message size is read from the env file
*/
func TestConfigMaxMessageSizeFromEnv(t *testing.T) {
	config, err := (&Configuration{}).Init("../../.env.tests")

	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	if config.RelayMaxMessageSize != 512000 {
		t.Fatalf("RelayMaxMessageSize should be 512000 but got %v", config.RelayMaxMessageSize)
	}
}

/*
TestConfigMaxMessageSizeDefault tests that a missing, zero, negative or invalid max message size falls back to the default
*/
func TestConfigMaxMessageSizeDefault(t *testing.T) {
	for _, value := range []string{"", "0", "-5", "abc"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("RELAY_MAX_MESSAGE_SIZE", value)

			config, err := (&Configuration{}).Init("../../.env.tests")
			if err != nil {
				t.Fatalf("Error on NewConfiguration: %v", err)
			}

			if config.RelayMaxMessageSize != 512*1024 {
				t.Fatalf("RelayMaxMessageSize should default to %v but got %v", 512*1024, config.RelayMaxMessageSize)
			}
		})
	}
}

/*
TestConfigMessageRateDefaults tests the default rate limiter settings when nothing is configured
*/
func TestConfigMessageRateDefaults(t *testing.T) {
	t.Setenv("RELAY_MESSAGE_RATE", "")
	t.Setenv("RELAY_MESSAGE_BURST", "")

	config, err := (&Configuration{}).Init("../../.env.tests")
	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	if config.RelayMessageRate != 20 {
		t.Fatalf("RelayMessageRate should default to 20 but got %v", config.RelayMessageRate)
	}

	if config.RelayMessageBurst != 50 {
		t.Fatalf("RelayMessageBurst should default to 50 but got %v", config.RelayMessageBurst)
	}
}

/*
TestConfigMessageRateFromEnv tests that the rate limiter settings are read from the environment
*/
func TestConfigMessageRateFromEnv(t *testing.T) {
	t.Setenv("RELAY_MESSAGE_RATE", "2.5")
	t.Setenv("RELAY_MESSAGE_BURST", "7")

	config, err := (&Configuration{}).Init("../../.env.tests")
	if err != nil {
		t.Fatalf("Error on NewConfiguration: %v", err)
	}

	if config.RelayMessageRate != 2.5 {
		t.Fatalf("RelayMessageRate should be 2.5 but got %v", config.RelayMessageRate)
	}

	if config.RelayMessageBurst != 7 {
		t.Fatalf("RelayMessageBurst should be 7 but got %v", config.RelayMessageBurst)
	}
}

/*
TestConfigMessageRateInvalid tests that invalid rate limiter settings fall back to the defaults
*/
func TestConfigMessageRateInvalid(t *testing.T) {
	for _, value := range []string{"0", "-1", "abc"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("RELAY_MESSAGE_RATE", value)
			t.Setenv("RELAY_MESSAGE_BURST", value)

			config, err := (&Configuration{}).Init("../../.env.tests")
			if err != nil {
				t.Fatalf("Error on NewConfiguration: %v", err)
			}

			if config.RelayMessageRate != 20 {
				t.Fatalf("RelayMessageRate should default to 20 but got %v", config.RelayMessageRate)
			}

			if config.RelayMessageBurst != 50 {
				t.Fatalf("RelayMessageBurst should default to 50 but got %v", config.RelayMessageBurst)
			}
		})
	}
}
