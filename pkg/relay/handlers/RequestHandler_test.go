package handlers

import (
	"encoding/json"
	"testing"
)

/*
TestRequestHandler_IDRequired tests that the RequestHandler returns is an error string when checking without id
*/
func TestRequestHandler_IDRequired(t *testing.T) {
	var msg = make([]json.RawMessage, 3)
	msg[0], _ = json.Marshal("REQ")
	msg[1], _ = json.Marshal("")

	rh := RequestHandler{Ctx: nil, Ws: nil, Req: msg}

	outputString := rh.Handle()

	if outputString != "REQ has no <id>" {
		t.Fatalf("Handle does not properly check id, result was: %v", outputString)
	}
}

/*
TestRequestHandler_FilterDecodeError tests that the RequestHandler properly returns an error string on decode error
*/
func TestRequestHandler_FilterDecodeError(t *testing.T) {
	var msg = make([]json.RawMessage, 3)
	msg[0], _ = json.Marshal("REQ")
	msg[1], _ = json.Marshal("RandomID")
	msg[2], _ = json.Marshal("")

	rh := RequestHandler{Ctx: nil, Ws: nil, Req: msg}

	outputString := rh.Handle()

	if outputString != "failed to decode filter" {
		t.Fatalf("Handle does not catch invalid filter query, result was: %v", outputString)
	}
}
