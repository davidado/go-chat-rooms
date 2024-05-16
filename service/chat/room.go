package chat

import (
	"log"
)

// PubSubber is an interface for subscribing to a room and
// publishing messages to it.
type PubSubber interface {
	Publish(topic string, m Message) error
	SubscribeAndBroadcast(room *Room)
}

// Room maintains the set of active clients and broadcasts messages to them.
type Room struct {
	// Name of the room.
	Name string

	// Clients in the room.
	Clients *Clients

	// PubSubber for the room.
	Pubsub PubSubber

	// Inbound messages from the clients.
	broadcast chan Message

	// register requests from clients.
	register chan *Client

	// unregister requests from clients.
	unregister chan *Client

	// Done channel lets the pubsub SubscribeAndBroadcast
	// goroutine know to close.
	// Don't use a done channel in pubsub since it is
	// only initialized once in main.
	Done chan struct{}
}

// NewRoom creates a new room.
func NewRoom(name string, clients *Clients, pubsub PubSubber) *Room {
	return &Room{
		Name:       name,
		Clients:    clients,
		Pubsub:     pubsub,
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Done:       make(chan struct{}),
	}
}

// Run runs the room.
func (r *Room) Run(rooms *Rooms) {
	defer func() {
		// Unsubscribe from the room and remove from the
		// room list if there are no clients left.
		r.closeIfEmpty(rooms)
	}()

	for {
		select {
		case c := <-r.register:
			// Add the client to the room.
			r.Clients.Add(c)

			// Broadcast a message to the room that a new client has entered.
			// If the client is the first one in the room, don't broadcast
			// through the pubsub system since it hasn't fully initialized yet.
			// Use the local broadcast channel instead.
			r.broadcastNewClientEntry(c)
		case c := <-r.unregister:
			// Remove the client from the room.
			r.Clients.Remove(c)

			// Broadcast a message to the room that a client has left.
			r.broadcastClientExit(c)

			// End this goroutine when this client leaves.
			return
		case message := <-r.broadcast:
			// Broadcast the message to the room.
			err := r.Pubsub.Publish(r.Name, message)
			if err != nil {
				log.Println("error publishing message:", err)
			}
		}
	}
}

func (r *Room) closeIfEmpty(rooms *Rooms) {
	if r.Clients.NumClients() > 0 {
		return
	}
	close(r.Done)
	rooms.Remove(r.Name)
}

func (r *Room) broadcastClientExit(c *Client) {
	m := NewMessage(r.Name, c.name, []byte(""), LeaveAction)
	err := r.Pubsub.Publish(r.Name, m)
	if err != nil {
		log.Println("error publishing message:", err)
	}
}

func (r *Room) broadcastNewClientEntry(c *Client) {
	m := NewMessage(r.Name, c.name, []byte(""), JoinAction)

	if r.Clients.NumClients() == 1 {
		// If the client is the first one in the room,
		// there is no need to broadcast to the pubsub
		// topic so just use the local broadcast channel.
		r.Broadcast(m)
	} else {
		err := r.Pubsub.Publish(r.Name, m)
		if err != nil {
			log.Println("error publishing message:", err)
		}
	}
}

// Broadcast sends a message to all clients in the room.
func (r *Room) Broadcast(m Message) {
	for c := range r.Clients.GetClients() {
		select {
		case c.send <- m:
		default:
			r.Clients.Remove(c)
			log.Println(c.name, "left the room")
		}
	}
}
