package logging

import (
	"artio-relay/pkg/storage/adapter"
	"fmt"
	"log"
	"net/http"
)

type Logger struct {
	LogAdapter adapter.LogAdapter
}

/*
getIP fetches the correct IP address from a given request
*/
func getIP(r *http.Request) string {
	var ip string
	if realIP := r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP // possible to be multiple comma separated
	} else if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	return ip
}

/*
LogConnect creates a log entry for Connect
*/
func (l *Logger) LogConnect(ip string) {
	log.Printf("connected from %s", ip)
	_, _ = l.LogAdapter.Create(ip, "CONN", "connected")
}

/*
LogDisconnect creates a log entry for LogDisconnect
*/
func (l *Logger) LogDisconnect(ip string) {
	log.Printf("disconnected from %s\n", ip)
	_, _ = l.LogAdapter.Create(ip, "CONN", "disconnected")
}

/*
LogPing creates a log entry for LogPing
*/
func (l *Logger) LogPing(ip string) {
	log.Printf("pinging for %s", ip)
	_, _ = l.LogAdapter.Create(ip, "PING", "executing ping")
}

/*
LogNIP11 create a log entry for a NIP 11 request
*/
func (l *Logger) LogNIP11(ip string) {
	log.Printf("handling NIP-11 request for %s", ip)
	_, _ = l.LogAdapter.Create(ip, "NIP11", "NIP-11 Request")
}

func (l *Logger) LogRequest(typ string, content string, ip string) {
	fmt.Println("message:", string(content))
	_, _ = l.LogAdapter.Create(ip, typ, content)
}

func (l *Logger) LogHandling(typ string, content string, ip string) {
	_, _ = l.LogAdapter.Create(ip, typ, content)
}

var ArtioLogger = Logger{adapter.LogAdapter{}}
