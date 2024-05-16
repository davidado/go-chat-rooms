// Package chat provides a chat room implementation.
package chat

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/davidado/go-chat-rooms/pkg/auth"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the room.
type Client struct {
	// The client's name.
	name string

	// The room the client is in.
	room *Room

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan Message
}

// NewClient creates a new client.
func NewClient(name string, room *Room, conn *websocket.Conn) *Client {
	return &Client{name: name, room: room, conn: conn, send: make(chan Message, 10)}
}

// ReadSocket reads messages from the websocket connection then broadcasts
// it to a room through a channel.
//
// A ReadSocket goroutine is started for each connection. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadSocket() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		msg := NewMessage(c.room.Name, c.name, message, MessageAction)
		c.room.broadcast <- msg
	}
}

// WriteSocket writes messages posted from a room to the websocket connection.
//
// A WriteSocket goroutine is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WriteSocket() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			// Don't send the message to the sender
			// that sent the message. Let that sender's
			// UI handle displaying the message to them.
			if message.Sender == c.name && message.Action == MessageAction {
				continue
			}

			// Only send the participant list to the client
			// that just joined the room. The existing clients
			// in the room update their existing participant lists
			// by tracking the join and leave actions.
			if message.Sender == c.name && message.Action == JoinAction {
				message.SetParticipants(c.room.Clients)
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The room closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			b, err := message.JSON()
			if err != nil {
				continue
			}
			w.Write(b)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				msg := <-c.send
				w.Write(msg.Message)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(room *Room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	username := auth.GetUsernameFromContext(r.Context())
	client := NewClient(username, room, conn)
	client.room.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WriteSocket()
	go client.ReadSocket()
}
