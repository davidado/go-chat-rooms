package chat

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

// ActionType is the type of action.
type ActionType string

// Action types.
const (
	JoinAction    ActionType = "joined"
	LeaveAction   ActionType = "left"
	MessageAction ActionType = "message"
)

// Message is a message sent from a client to a room
// along with its meta details.
type Message struct {
	// Room name.
	RoomName string

	// Sender name.
	Sender string

	// Message.
	Message []byte

	// Participants in the room.
	Participants map[string]bool

	// Action type.
	Action ActionType
}

// NewMessage creates a new message.
func NewMessage(roomName, sender string, message []byte, action ActionType) Message {
	return Message{
		RoomName: roomName,
		Sender:   sender,
		Message:  message,
		Action:   action,
	}
}

// Marshal serializes the message.
func (m *Message) Marshal() ([]byte, error) {
	b, err := msgpack.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Unmarshal deserializes the message.
func (m *Message) Unmarshal(b []byte) error {
	err := msgpack.Unmarshal(b, m)
	if err != nil {
		return err
	}
	return nil
}

// SetParticipants includes the participant list in the message.
func (m *Message) SetParticipants(clients *Clients) {
	participants := make(map[string]bool)
	cs := clients.GetClients()
	for c := range cs {
		participants[c.name] = true
	}
	m.Participants = participants
}

// JSON serializes the message to JSON.
func (m *Message) JSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Sender       string          `json:"sender"`
		Message      string          `json:"message"`
		Participants map[string]bool `json:"participants,omitempty"`
		Action       string          `json:"action"`
	}{
		Sender:       m.Sender,
		Message:      string(m.Message),
		Participants: m.Participants,
		Action:       string(m.Action),
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}
