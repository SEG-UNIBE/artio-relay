package handlers

import (
	"context"
	"encoding/json"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/webSocket"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip42"
)

const AuthContextKey = iota

/*
AuthenticationHandler handles messages of type AUTH that are associated with NIP-42
*/
type AuthenticationHandler struct {
	Ctx *context.Context
	Ws  *webSocket.WebSocket
	Req []json.RawMessage
}

func (ah *AuthenticationHandler) Handle() string {
	if !config.Config.SupportNIP42 {
		return ""
	}

	// we do indeed support it
	var evt nostr.Event
	if err := json.Unmarshal(ah.Req[1], &evt); err != nil {
		return "failed to decode auth event: " + err.Error()
	}
	if pubkey, ok := nip42.ValidateAuthEvent(&evt, ah.Ws.Challenge, ah.Ws.ServiceURL); ok {
		ah.Ws.Authenticated = pubkey
		ctx := context.WithValue(*ah.Ctx, AuthContextKey, pubkey)
		ah.Ctx = &ctx
		_ = ah.Ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: true})
	} else {
		_ = ah.Ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: false, Reason: "error: failed to authenticate"})
	}
	return ""
}
