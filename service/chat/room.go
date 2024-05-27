package chat

import (
	"context"
	"log"
)

// PubSubber is an interface for subscribing to a room and
// publishing messages to it.
type PubSubber interface {
	Publish(ctx context.Context, topic string, msg []byte) error
	Subscribe(ctx context.Context, topic string, payload chan []byte)
}

// Room maintains the set of active clients and broadcasts messages to them.
type Room struct {
	// The context of the room.
	ctx context.Context

	// The cancel function for the room context.
	cancel context.CancelFunc

	// name of the room.
	name string

	// The pubsub system for the room.
	pubsub PubSubber

	// clients in the room.
	clients *Clients

	// Messages from the pubsub subscriber.
	pbPayload chan []byte

	// Inbound messages from the clients.
	clientMsg chan Message

	// register requests from clients.
	register chan *Client

	// unregister requests from clients.
	unregister chan *Client
}

// NewRoom creates a new room.
func NewRoom(ctx context.Context, cancel context.CancelFunc, name string, pubsub PubSubber) *Room {
	return &Room{
		ctx:        ctx,
		cancel:     cancel,
		name:       name,
		pubsub:     pubsub,
		clients:    NewClients(name),
		pbPayload:  make(chan []byte),
		clientMsg:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Subscribe subscribes to the pubsub topic for each room.
func (r *Room) Subscribe() {
	go r.pubsub.Subscribe(r.ctx, r.name, r.pbPayload)
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
			r.clients.Add(c)

			// Broadcast a message to the room that a new client has entered.
			// If the client is the first one in the room, don't broadcast
			// through the pubsub system since it hasn't fully initialized yet.
			// Use the local broadcast channel instead.
			r.broadcastNewClientEntry(r.ctx, c, r.pubsub)
		case c := <-r.unregister:
			// Remove the client from the room.
			r.clients.Remove(c)

			// Broadcast a message to the room that a client has left.
			r.broadcastClientExit(r.ctx, c, r.pubsub)

			// End this goroutine when this client leaves.
			return
		case msg := <-r.clientMsg:
			// Publish the client message to the pubsub topic.
			b, err := msg.Marshal()
			if err != nil {
				log.Println("Run - error marshalling message:", err)
			}
			err = r.pubsub.Publish(r.ctx, r.name, b)
			if err != nil {
				log.Println("Run - error publishing message:", err)
			}
		case b := <-r.pbPayload:
			// Broadcast the message from the pubsub subscriber to
			// each client in the room.
			r.Broadcast(b)
		}
	}
}

func (r *Room) closeIfEmpty(rooms *Rooms) {
	if r.clients.NumClients() > 0 {
		return
	}

	r.cancel()
	rooms.Remove(r.name)
}

func (r *Room) broadcastClientExit(ctx context.Context, c *Client, pb PubSubber) {
	if r.clients.NumClients() == 0 {
		// Don't broadcast an exit message to the room
		// for the last client that leaves.
		// Not only is this unnecessary, but it also
		// conflicts with the ctx.Done() check in the
		// Subscribe method of pubsub resulting in
		// non-deterministic behavior and a leaky
		// goroutine. The Subscribe method only either
		// detected the broadcasted exit message or
		// the ctx.Done() signal, but not both.
		// Ask me how many hours it took to figure
		// out why the Subscribe goroutine only
		// exited sometimes.
		return
	}

	m := NewMessage(r.name, c.name, []byte(""), LeaveAction)
	b, err := m.Marshal()
	if err != nil {
		log.Println("broadcastClientExit - error marshalling message:", err)
	}
	err = pb.Publish(ctx, r.name, b)
	if err != nil {
		log.Println("broadcastClientExit - error publishing message:", err)
	}
}

func (r *Room) broadcastNewClientEntry(ctx context.Context, c *Client, pb PubSubber) {
	m := NewMessage(r.name, c.name, []byte(""), JoinAction)
	if r.clients.NumClients() == 1 {
		// If the client is the first one in the room,
		// there is no need to broadcast to the pubsub
		// topic so just use the local broadcast channel.
		r.BroadcastMessage(m)
	} else {
		b, err := m.Marshal()
		if err != nil {
			log.Println("broadcastNewClientEntry - error marshalling message:", err)
		}
		err = pb.Publish(ctx, r.name, b)
		if err != nil {
			log.Println("broadcastNewClientEntry - error publishing message:", err)
		}
	}
}

// BroadcastMessage sends a message to all clients in the room.
func (r *Room) BroadcastMessage(m Message) {
	for c := range r.clients.GetClients() {
		select {
		case c.send <- m:
		}
	}
}

// Broadcast sends a message to all clients in the room.
func (r *Room) Broadcast(msg []byte) {
	m := Message{}
	err := m.Unmarshal(msg)
	if err != nil {
		log.Print("Broadcast - error unmarshalling message:", err)
		return
	}

	r.BroadcastMessage(m)
}
