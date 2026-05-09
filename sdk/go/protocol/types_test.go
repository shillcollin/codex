package protocol

import (
	"encoding/json"
	"testing"
)

func TestResourceContentUnmarshalText(t *testing.T) {
	var content ResourceContent
	if err := json.Unmarshal([]byte(`{"uri":"file:///note.txt","text":"hello"}`), &content); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if content.URI != "file:///note.txt" || content.Text == nil || *content.Text != "hello" || content.Blob != nil {
		t.Fatalf("unexpected decoded content: %+v", content)
	}
}

func TestThreadRealtimeStartTransportMarshalWebrtc(t *testing.T) {
	transport := NewThreadRealtimeWebrtcTransport("offer-sdp")
	data, err := json.Marshal(transport)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if string(data) != `{"type":"webrtc","sdp":"offer-sdp"}` {
		t.Fatalf("unexpected encoded transport: %s", data)
	}
}

func TestThreadRealtimeStartTransportRejectsInvalidWebsocketShape(t *testing.T) {
	var transport ThreadRealtimeStartTransport
	err := json.Unmarshal([]byte(`{"type":"websocket","sdp":"unexpected"}`), &transport)
	if err == nil {
		t.Fatal("expected websocket transport to reject sdp")
	}
}
