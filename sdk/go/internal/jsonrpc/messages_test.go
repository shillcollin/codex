package jsonrpc

import "testing"

func TestRequestIDRoundTripsStringAndInteger(t *testing.T) {
	stringID := NewStringRequestID("req-1")
	if stringID.String() != "req-1" || stringID.Key() != "s:req-1" {
		t.Fatalf("unexpected string id: %#v", stringID)
	}

	integerID := NewIntegerRequestID(42)
	if integerID.String() != "42" || integerID.Key() != "i:42" {
		t.Fatalf("unexpected integer id: %#v", integerID)
	}
}

func TestParseEnvelopeWithIntegerID(t *testing.T) {
	env, err := ParseEnvelope([]byte(`{"id":42,"result":{"ok":true}}`))
	if err != nil {
		t.Fatalf("ParseEnvelope returned error: %v", err)
	}
	if env.ID == nil {
		t.Fatal("expected envelope id")
	}
	if env.ID.String() != "42" || env.ID.Key() != "i:42" {
		t.Fatalf("unexpected parsed id: %#v", env.ID)
	}
}
