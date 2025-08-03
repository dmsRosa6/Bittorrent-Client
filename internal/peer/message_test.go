package peer

import (
	"bytes"
	"testing"
)

func Test_MesssageSerialize_OK(t *testing.T){
	mId := messageID(1)
	payload := "stuff"

	expected := make([]byte, len(payload) + 5)



	m := Message{ID: mId, Payload: []byte(payload)}

	buf := m.Serialize()

	if len(expected) != len(buf) {
		t.Fatalf("lengths do not match. expected %d, got %d", len(expected), len(buf))
	}

	if buf[4] != byte(mId) {
		t.Errorf("message ID mismatch: expected %d, got %d", mId, buf[4])
	}

	expectedPayload := []byte(payload)
	if !bytes.Equal(buf[5:], expectedPayload) {
		t.Errorf("payload mismatch: expected %q, got %q", expectedPayload, buf[5:])
	}
}

func Test_MessageRead_OK(t *testing.T) {
    // Build a raw message buffer
    payload := []byte("hello")
    id := messageID(2)
    msg := &Message{ID: id, Payload: payload}
    buf := msg.Serialize()

    parsed, err := ReadMessage(buf)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if parsed.ID != id {
        t.Errorf("expected ID %d, got %d", id, parsed.ID)
    }

    if !bytes.Equal(parsed.Payload, payload) {
        t.Errorf("payload mismatch: expected %q, got %q", payload, parsed.Payload)
    }
}

func Test_MessageSerializeThenRead_OK(t *testing.T) {
    original := &Message{
        ID:      messageID(5),
        Payload: []byte("this is test data"),
    }

    buf := original.Serialize()
    result, err := ReadMessage(buf)
    if err != nil {
        t.Fatalf("failed to parse serialized message: %v", err)
    }

    if result.ID != original.ID {
        t.Errorf("ID mismatch: expected %d, got %d", original.ID, result.ID)
    }

    if !bytes.Equal(result.Payload, original.Payload) {
        t.Errorf("Payload mismatch: expected %q, got %q", original.Payload, result.Payload)
    }
}
