// Package protocol implements the Volcengine WebSocket binary protocol.
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// MsgType represents message type.
type MsgType uint8

const (
	MsgInvalid MsgType = 0
	MsgFullClientRequest MsgType = 0b1
	MsgAudioOnlyClient MsgType = 0b10
	MsgFullServerResponse MsgType = 0b1001
	MsgAudioOnlyServer MsgType = 0b1011
	MsgFrontEndResultServer MsgType = 0b1100
	MsgError MsgType = 0b1111

	// Alias
	MsgServerACK = MsgAudioOnlyServer
)

func (m MsgType) String() string {
	switch m {
	case MsgInvalid:
		return "Invalid"
	case MsgFullClientRequest:
		return "FullClientRequest"
	case MsgAudioOnlyClient:
		return "AudioOnlyClient"
	case MsgFullServerResponse:
		return "FullServerResponse"
	case MsgAudioOnlyServer:
		return "AudioOnlyServer"
	case MsgFrontEndResultServer:
		return "FrontEndResultServer"
	case MsgError:
		return "Error"
	default:
		return fmt.Sprintf("MsgType(%d)", m)
	}
}

// MsgTypeFlag represents message type flags.
type MsgTypeFlag uint8

const (
	FlagNoSeq MsgTypeFlag = 0        // Non-terminal packet with no sequence
	FlagPositiveSeq MsgTypeFlag = 0b1  // Non-terminal packet with sequence > 0
	FlagLastNoSeq MsgTypeFlag = 0b10   // Last packet with no sequence
	FlagNegativeSeq MsgTypeFlag = 0b11 // Last packet with sequence < 0
	FlagWithEvent MsgTypeFlag = 0b100  // Payload contains event number
)

// Version represents protocol version.
type Version uint8

const (
	Version1 Version = 1
	Version2 Version = 2
	Version3 Version = 3
	Version4 Version = 4
)

// HeaderSize represents header size.
type HeaderSize uint8

const (
	HeaderSize4 HeaderSize = 1
	HeaderSize8 HeaderSize = 2
	HeaderSize12 HeaderSize = 3
	HeaderSize16 HeaderSize = 4
)

// Serialization represents serialization method.
type Serialization uint8

const (
	SerialRaw Serialization = 0
	SerialJSON Serialization = 0b1
	SerialThrift Serialization = 0b11
	SerialCustom Serialization = 0b1111
)

// Compression represents compression method.
type Compression uint8

const (
	CompressNone Compression = 0
	CompressGzip Compression = 0b1
	CompressCustom Compression = 0b1111
)

// EventType represents event type.
type EventType int32

const (
	EventNone EventType = 0

	// 1 ~ 49 Upstream Connection events
	EventStartConnection EventType = 1
	EventStartTask EventType = 1 // Alias
	EventFinishConnection EventType = 2
	EventFinishTask EventType = 2 // Alias

	// 50 ~ 99 Downstream Connection events
	EventConnectionStarted EventType = 50
	EventTaskStarted EventType = 50 // Alias
	EventConnectionFailed EventType = 51
	EventTaskFailed EventType = 51 // Alias
	EventConnectionFinished EventType = 52
	EventTaskFinished EventType = 52 // Alias

	// 100 ~ 149 Upstream Session events
	EventStartSession EventType = 100
	EventCancelSession EventType = 101
	EventFinishSession EventType = 102

	// 150 ~ 199 Downstream Session events
	EventSessionStarted EventType = 150
	EventSessionCanceled EventType = 151
	EventSessionFinished EventType = 152
	EventSessionFailed EventType = 153
	EventUsageResponse EventType = 154
	EventChargeData EventType = 154 // Alias

	// 200 ~ 249 Upstream general events
	EventTaskRequest EventType = 200
	EventUpdateConfig EventType = 201

	// 250 ~ 299 Downstream general events
	EventAudioMuted EventType = 250

	// 300 ~ 349 Upstream TTS events
	EventSayHello EventType = 300

	// 350 ~ 399 Downstream TTS events
	EventTTSSentenceStart EventType = 350
	EventTTSSentenceEnd EventType = 351
	EventTTSResponse EventType = 352
	EventTTSEnded EventType = 359
	EventPodcastRoundStart EventType = 360
	EventPodcastRoundResponse EventType = 361
	EventPodcastRoundEnd EventType = 362
)

func (e EventType) String() string {
	switch e {
	case EventNone:
		return "None"
	case EventStartConnection:
		return "StartConnection"
	case EventFinishConnection:
		return "FinishConnection"
	case EventConnectionStarted:
		return "ConnectionStarted"
	case EventConnectionFailed:
		return "ConnectionFailed"
	case EventConnectionFinished:
		return "ConnectionFinished"
	case EventStartSession:
		return "StartSession"
	case EventCancelSession:
		return "CancelSession"
	case EventFinishSession:
		return "FinishSession"
	case EventSessionStarted:
		return "SessionStarted"
	case EventSessionCanceled:
		return "SessionCanceled"
	case EventSessionFinished:
		return "SessionFinished"
	case EventSessionFailed:
		return "SessionFailed"
	case EventTaskRequest:
		return "TaskRequest"
	case EventTTSSentenceStart:
		return "TTSSentenceStart"
	case EventTTSSentenceEnd:
		return "TTSSentenceEnd"
	case EventTTSResponse:
		return "TTSResponse"
	default:
		return fmt.Sprintf("EventType(%d)", e)
	}
}

// Message represents a protocol message.
type Message struct {
	Version      Version
	HeaderSize   HeaderSize
	Type         MsgType
	Flag         MsgTypeFlag
	Serialization Serialization
	Compression  Compression

	Event     EventType
	SessionID string
	ConnectID string
	Sequence  int32
	ErrorCode uint32

	Payload []byte
}

// VolcengineProtocol handles protocol operations.
type VolcengineProtocol struct{}

// NewVolcengineProtocol creates a new protocol handler.
func NewVolcengineProtocol() *VolcengineProtocol {
	return &VolcengineProtocol{}
}

// Marshal serializes a message to bytes.
func (p *VolcengineProtocol) Marshal(msg *Message) []byte {
	buf := new(bytes.Buffer)

	// Write header (byte 0: version + header_size)
	binary.Write(buf, binary.BigEndian, uint8((msg.Version<<4)|msg.HeaderSize))

	// Write header (byte 1: type + flag)
	binary.Write(buf, binary.BigEndian, uint8((msg.Type<<4)|msg.Flag))

	// Write header (byte 2: serialization + compression)
	binary.Write(buf, binary.BigEndian, uint8((msg.Serialization<<4)|msg.Compression))

	// Calculate header padding
	headerBytes := 4 * int(msg.HeaderSize)
	padding := headerBytes - 3
	for i := 0; i < padding; i++ {
		binary.Write(buf, binary.BigEndian, uint8(0))
	}

	// Write event if flag indicates
	if msg.Flag == FlagWithEvent {
		binary.Write(buf, binary.BigEndian, msg.Event)

		// Write session ID (skip for connection events)
		if !isConnectionEvent(msg.Event) {
			sessionBytes := []byte(msg.SessionID)
			binary.Write(buf, binary.BigEndian, uint32(len(sessionBytes)))
			if len(sessionBytes) > 0 {
				buf.Write(sessionBytes)
			}
		}
	}

	// Write sequence for certain message types
	if needsSequence(msg.Type, msg.Flag) {
		binary.Write(buf, binary.BigEndian, msg.Sequence)
	}

	// Write error code for error messages
	if msg.Type == MsgError {
		binary.Write(buf, binary.BigEndian, msg.ErrorCode)
	}

	// Write payload
	binary.Write(buf, binary.BigEndian, uint32(len(msg.Payload)))
	buf.Write(msg.Payload)

	return buf.Bytes()
}

// Unmarshal deserializes bytes to a message.
func (p *VolcengineProtocol) Unmarshal(reader io.Reader) (*Message, error) {
	msg := &Message{
		Version:      Version1,
		HeaderSize:   HeaderSize4,
		Serialization: SerialJSON,
		Compression:  CompressNone,
	}

	// Read byte 0: version + header_size
	var b0 uint8
	if err := binary.Read(reader, binary.BigEndian, &b0); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	msg.Version = Version(b0 >> 4)
	msg.HeaderSize = HeaderSize(b0 & 0b00001111)

	// Read byte 1: type + flag
	var b1 uint8
	if err := binary.Read(reader, binary.BigEndian, &b1); err != nil {
		return nil, fmt.Errorf("failed to read type: %w", err)
	}
	msg.Type = MsgType(b1 >> 4)
	msg.Flag = MsgTypeFlag(b1 & 0b00001111)

	// Read byte 2: serialization + compression
	var b2 uint8
	if err := binary.Read(reader, binary.BigEndian, &b2); err != nil {
		return nil, fmt.Errorf("failed to read serialization: %w", err)
	}
	msg.Serialization = Serialization(b2 >> 4)
	msg.Compression = Compression(b2 & 0b00001111)

	// Skip header padding
	headerBytes := 4 * int(msg.HeaderSize)
	padding := headerBytes - 3
	for i := 0; i < padding; i++ {
		var skip uint8
		binary.Read(reader, binary.BigEndian, &skip)
	}

	// Read sequence for certain message types
	if needsSequence(msg.Type, msg.Flag) {
		if err := binary.Read(reader, binary.BigEndian, &msg.Sequence); err != nil {
			return nil, fmt.Errorf("failed to read sequence: %w", err)
		}
	}

	// Read error code for error messages
	if msg.Type == MsgError {
		if err := binary.Read(reader, binary.BigEndian, &msg.ErrorCode); err != nil {
			return nil, fmt.Errorf("failed to read error code: %w", err)
		}
	}

	// Read event if flag indicates
	if msg.Flag == FlagWithEvent {
		if err := binary.Read(reader, binary.BigEndian, &msg.Event); err != nil {
			return nil, fmt.Errorf("failed to read event: %w", err)
		}

		// Read session ID (skip for connection events)
		if !isConnectionEvent(msg.Event) {
			var sessionLen uint32
			if err := binary.Read(reader, binary.BigEndian, &sessionLen); err != nil {
				return nil, fmt.Errorf("failed to read session length: %w", err)
			}
			if sessionLen > 0 {
				sessionBytes := make([]byte, sessionLen)
				if _, err := io.ReadFull(reader, sessionBytes); err != nil {
					return nil, fmt.Errorf("failed to read session ID: %w", err)
				}
				msg.SessionID = string(sessionBytes)
			}
		}

		// Read connect ID for certain events
		if isConnectIDEvent(msg.Event) {
			var connectLen uint32
			if err := binary.Read(reader, binary.BigEndian, &connectLen); err != nil {
				return nil, fmt.Errorf("failed to read connect length: %w", err)
			}
			if connectLen > 0 {
				connectBytes := make([]byte, connectLen)
				if _, err := io.ReadFull(reader, connectBytes); err != nil {
					return nil, fmt.Errorf("failed to read connect ID: %w", err)
				}
				msg.ConnectID = string(connectBytes)
			}
		}
	}

	// Read payload
	var payloadLen uint32
	if err := binary.Read(reader, binary.BigEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %w", err)
	}
	if payloadLen > 0 {
		msg.Payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(reader, msg.Payload); err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	}

	return msg, nil
}

// Helper functions

func isConnectionEvent(event EventType) bool {
	return event == EventStartConnection ||
		event == EventFinishConnection ||
		event == EventConnectionStarted ||
		event == EventConnectionFailed
}

func isConnectIDEvent(event EventType) bool {
	return event == EventConnectionStarted ||
		event == EventConnectionFailed ||
		event == EventConnectionFinished
}

func needsSequence(msgType MsgType, flag MsgTypeFlag) bool {
	if msgType == MsgError {
		return false
	}
	if msgType == MsgFullClientRequest ||
		msgType == MsgFullServerResponse ||
		msgType == MsgFrontEndResultServer ||
		msgType == MsgAudioOnlyClient ||
		msgType == MsgAudioOnlyServer {
		return flag == FlagPositiveSeq || flag == FlagNegativeSeq
	}
	return false
}

// String returns a string representation of the message.
func (msg *Message) String() string {
	payloadLen := len(msg.Payload)
	if msg.Type == MsgAudioOnlyServer || msg.Type == MsgAudioOnlyClient {
		if msg.Flag == FlagPositiveSeq || msg.Flag == FlagNegativeSeq {
			return fmt.Sprintf("MsgType: %s, Event: %s, Seq: %d, PayloadLen: %d",
				msg.Type, msg.Event, msg.Sequence, payloadLen)
		}
		return fmt.Sprintf("MsgType: %s, Event: %s, PayloadLen: %d",
			msg.Type, msg.Event, payloadLen)
	}
	if msg.Type == MsgError {
		return fmt.Sprintf("MsgType: %s, Event: %s, ErrorCode: %d, Payload: %s",
			msg.Type, msg.Event, msg.ErrorCode, string(msg.Payload))
	}
	if msg.Flag == FlagPositiveSeq || msg.Flag == FlagNegativeSeq {
		return fmt.Sprintf("MsgType: %s, Event: %s, Seq: %d, Payload: %s",
			msg.Type, msg.Event, msg.Sequence, string(msg.Payload))
	}
	return fmt.Sprintf("MsgType: %s, Event: %s, Payload: %s",
		msg.Type, msg.Event, string(msg.Payload))
}