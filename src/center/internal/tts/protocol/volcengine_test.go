package protocol

import (
	"bytes"
	"testing"
)

func TestMarshalBasicMessage(t *testing.T) {
	proto := NewVolcengineProtocol()

	msg := &Message{
		Version:      Version1,
		HeaderSize:   HeaderSize4,
		Type:         MsgFullClientRequest,
		Flag:         FlagWithEvent,
		Serialization: SerialJSON,
		Compression:  CompressNone,
		Event:        EventStartConnection,
		Payload:      []byte("{}"),
	}

	data := proto.Marshal(msg)

	// Basic sanity check: should have at least header + event + payload
	if len(data) < 12 {
		t.Errorf("Message too short: got %d bytes", len(data))
	}

	// Check version and header size in first byte
	if data[0] != 0x14 { // version 1 << 4 | header_size 1
		t.Errorf("First byte mismatch: got 0x%02x, expected 0x14", data[0])
	}
}

func TestMarshalSessionMessage(t *testing.T) {
	proto := NewVolcengineProtocol()

	msg := &Message{
		Version:      Version1,
		HeaderSize:   HeaderSize4,
		Type:         MsgFullClientRequest,
		Flag:         FlagWithEvent,
		Serialization: SerialJSON,
		Compression:  CompressNone,
		Event:        EventStartSession,
		SessionID:    "test_session_123",
		Payload:      []byte("{\"test\": true}"),
	}

	data := proto.Marshal(msg)

	// Should include session ID length + session ID bytes
	if len(data) < 20 {
		t.Errorf("Message too short for session: got %d bytes", len(data))
	}
}

func TestUnmarshalBasicMessage(t *testing.T) {
	proto := NewVolcengineProtocol()

	// Create a message and marshal it
	original := &Message{
		Version:      Version1,
		HeaderSize:   HeaderSize4,
		Type:         MsgFullClientRequest,
		Flag:         FlagWithEvent,
		Serialization: SerialJSON,
		Compression:  CompressNone,
		Event:        EventStartConnection,
		Payload:      []byte("{}"),
	}

	data := proto.Marshal(original)
	reader := bytes.NewReader(data)

	msg, err := proto.Unmarshal(reader)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check fields
	if msg.Type != original.Type {
		t.Errorf("Type mismatch: got %d, expected %d", msg.Type, original.Type)
	}
	if msg.Flag != original.Flag {
		t.Errorf("Flag mismatch: got %d, expected %d", msg.Flag, original.Flag)
	}
	if msg.Event != original.Event {
		t.Errorf("Event mismatch: got %d, expected %d", msg.Event, original.Event)
	}
}

func TestUnmarshalAudioMessage(t *testing.T) {
	proto := NewVolcengineProtocol()

	// Create an audio-only server message
	original := &Message{
		Version:      Version1,
		HeaderSize:   HeaderSize4,
		Type:         MsgAudioOnlyServer,
		Flag:         FlagPositiveSeq,
		Serialization: SerialRaw,
		Compression:  CompressNone,
		Sequence:     1,
		Payload:      []byte{0x00, 0x01, 0x02, 0x03},
	}

	data := proto.Marshal(original)
	reader := bytes.NewReader(data)

	msg, err := proto.Unmarshal(reader)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if msg.Type != MsgAudioOnlyServer {
		t.Errorf("Type mismatch: got %s, expected AudioOnlyServer", msg.Type)
	}
	if msg.Sequence != 1 {
		t.Errorf("Sequence mismatch: got %d, expected 1", msg.Sequence)
	}
	if len(msg.Payload) != 4 {
		t.Errorf("Payload length mismatch: got %d, expected 4", len(msg.Payload))
	}
}

func TestUnmarshalSessionMessage(t *testing.T) {
	proto := NewVolcengineProtocol()

	original := &Message{
		Version:      Version1,
		HeaderSize:   HeaderSize4,
		Type:         MsgFullClientRequest,
		Flag:         FlagWithEvent,
		Serialization: SerialJSON,
		Compression:  CompressNone,
		Event:        EventStartSession,
		SessionID:    "session_abc",
		Payload:      []byte("{\"test\": \"data\"}"),
	}

	data := proto.Marshal(original)
	reader := bytes.NewReader(data)

	msg, err := proto.Unmarshal(reader)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if msg.SessionID != "session_abc" {
		t.Errorf("SessionID mismatch: got %s, expected session_abc", msg.SessionID)
	}
}

func TestMsgTypeString(t *testing.T) {
	tests := []struct {
		typ   MsgType
		want  string
	}{
		{MsgInvalid, "Invalid"},
		{MsgFullClientRequest, "FullClientRequest"},
		{MsgAudioOnlyServer, "AudioOnlyServer"},
		{MsgError, "Error"},
	}

	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("MsgType.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		event EventType
		want  string
	}{
		{EventNone, "None"},
		{EventStartConnection, "StartConnection"},
		{EventConnectionStarted, "ConnectionStarted"},
		{EventStartSession, "StartSession"},
		{EventTaskRequest, "TaskRequest"},
	}

	for _, tt := range tests {
		if got := tt.event.String(); got != tt.want {
			t.Errorf("EventType.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestMessageString(t *testing.T) {
	msg := &Message{
		Type:    MsgFullClientRequest,
		Event:   EventStartConnection,
		Payload: []byte("{}"),
	}

	str := msg.String()
	if !contains(str, "FullClientRequest") {
		t.Errorf("Message.String() missing type: %s", str)
	}
	if !contains(str, "StartConnection") {
		t.Errorf("Message.String() missing event: %s", str)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}