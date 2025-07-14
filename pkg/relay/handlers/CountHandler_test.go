package handlers

import (
	"encoding/json"
	"testing"
)

/*
TestCountHandler_IDRequired tests that the RequestHandler returns is an error string when checking without id
*/
func TestCountHandler_IDRequired(t *testing.T) {
	var msg = make([]json.RawMessage, 3)
	msg[0], _ = json.Marshal("REQ")
	msg[1], _ = json.Marshal("")

	rh := CountHandler{Ctx: nil, Ws: nil, Req: msg}

	outputString := rh.Handle()

	if outputString != "REQ has no <id>" {
		t.Fatalf("Handle does not properly check id, result was: %v", outputString)
	}
}
