package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PromStats struct {
	Nip11Handles      prometheus.Counter
	WebSocketUpgrades prometheus.Counter

	RelayMessagesIn  prometheus.Counter
	RelayMessagesOut prometheus.Counter

	RelayTypeCounter *prometheus.CounterVec
}

/*
Get returns the PromStats instance, initializing it if necessary
*/
func (ps *PromStats) Get() *PromStats {
	ps.Init()
	return ps
}

/*
Init initializes all Prometheus metrics
*/
func (ps *PromStats) Init() {

	ps.Nip11Handles = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_nip11_handles_total",
		Help: "The total number of NIP-11 handles",
	})

	ps.WebSocketUpgrades = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_websocket_upgrades_total",
		Help: "The total number of WebSocket upgrades",
	})

	ps.RelayMessagesIn = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relay_messages_in_total",
		Help: "The total number of messages received by the relay",
	})

	ps.RelayMessagesOut = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relay_messages_out_total",
		Help: "The total number of messages sent by the relay",
	})

	ps.RelayTypeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "relay_message_typ_total",
		Help: "The total number of messages received by type",
	}, []string{"type"})

	// default labels:
	var defaultValues = []string{"EVENT", "COUNT", "REQ", "CLOSE", "AUTH"}
	for _, v := range defaultValues {
		ps.RelayTypeCounter.WithLabelValues(v)
	}

}

/*
Nip11Handled increments the NIP-11 handles counter
*/
func Nip11Handled() {
	PromStatsInstance.Nip11Handles.Inc()
}

/*
WebSocketUpgraded increments the WebSocket upgrades counter
*/
func WebSocketUpgraded() {
	PromStatsInstance.WebSocketUpgrades.Inc()
}

/*
MessageIn records an incoming message of a given type
*/
func MessageIn(typ string) {
	go EventReceived(typ)
	PromStatsInstance.RelayMessagesIn.Inc()
}

/*
MessageOut records an outgoing message
*/
func MessageOut() {
	PromStatsInstance.RelayMessagesOut.Inc()
}

/*
EventReceived records an event of a given type
*/
func EventReceived(typ string) {
	PromStatsInstance.RelayTypeCounter.WithLabelValues(typ).Inc()
}

// PromStatsInstance make the singleton instance
var PromStatsInstance = (&PromStats{}).Get()
