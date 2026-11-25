package stats

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

/*
counterValue reads the current value of a counter
*/
func counterValue(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()
	var metric dto.Metric
	if err := counter.Write(&metric); err != nil {
		t.Fatalf("failed to read counter: %v", err)
	}
	return metric.GetCounter().GetValue()
}

/*
waitForValue polls a collector until it reaches the expected value, needed for counters incremented in goroutines
*/
func waitForValue(t *testing.T, counter prometheus.Counter, expected float64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if counterValue(t, counter) == expected {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected value %v but got %v", expected, counterValue(t, counter))
}

/*
TestNip11Handled tests that the NIP-11 counter is incremented
*/
func TestNip11Handled(t *testing.T) {
	before := counterValue(t, PromStatsInstance.Nip11Handles)
	Nip11Handled()

	if after := counterValue(t, PromStatsInstance.Nip11Handles); after != before+1 {
		t.Fatalf("expected %v but got %v", before+1, after)
	}
}

/*
TestWebSocketUpgraded tests that the websocket upgrade counter is incremented
*/
func TestWebSocketUpgraded(t *testing.T) {
	before := counterValue(t, PromStatsInstance.WebSocketUpgrades)
	WebSocketUpgraded()

	if after := counterValue(t, PromStatsInstance.WebSocketUpgrades); after != before+1 {
		t.Fatalf("expected %v but got %v", before+1, after)
	}
}

/*
TestMessageOut tests that the outgoing message counter is incremented
*/
func TestMessageOut(t *testing.T) {
	before := counterValue(t, PromStatsInstance.RelayMessagesOut)
	MessageOut()

	if after := counterValue(t, PromStatsInstance.RelayMessagesOut); after != before+1 {
		t.Fatalf("expected %v but got %v", before+1, after)
	}
}

/*
TestMessageIn tests that the incoming message counter and the per type counter are incremented
*/
func TestMessageIn(t *testing.T) {
	typeCounter := PromStatsInstance.RelayTypeCounter.WithLabelValues("REQ")
	beforeTotal := counterValue(t, PromStatsInstance.RelayMessagesIn)
	beforeType := counterValue(t, typeCounter)

	MessageIn("REQ")

	if after := counterValue(t, PromStatsInstance.RelayMessagesIn); after != beforeTotal+1 {
		t.Fatalf("expected %v but got %v", beforeTotal+1, after)
	}
	// the per type counter is incremented asynchronously
	waitForValue(t, typeCounter, beforeType+1)
}

/*
TestEventReceived tests that the per type counter is incremented for the given type only
*/
func TestEventReceived(t *testing.T) {
	eventCounter := PromStatsInstance.RelayTypeCounter.WithLabelValues("EVENT")
	closeCounter := PromStatsInstance.RelayTypeCounter.WithLabelValues("CLOSE")
	beforeEvent := counterValue(t, eventCounter)
	beforeClose := counterValue(t, closeCounter)

	EventReceived("EVENT")

	if after := counterValue(t, eventCounter); after != beforeEvent+1 {
		t.Fatalf("expected %v but got %v", beforeEvent+1, after)
	}
	if after := counterValue(t, closeCounter); after != beforeClose {
		t.Fatalf("CLOSE counter should not change, expected %v but got %v", beforeClose, after)
	}
}

/*
TestDefaultTypeLabels tests that the default message types are initialized on startup
*/
func TestDefaultTypeLabels(t *testing.T) {
	metrics := make(chan prometheus.Metric, 100)
	PromStatsInstance.RelayTypeCounter.Collect(metrics)
	close(metrics)

	if count := len(metrics); count < 5 {
		t.Fatalf("expected at least the 5 default type labels but got %v", count)
	}
}
