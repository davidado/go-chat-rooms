package chat

// Clients maintains the set of active clients in a room.
type Clients struct {
	// Clients registered in the room.
	clients map[*Client]bool

	// Room name
	roomName string
}

// NewClients creates a new client store.
func NewClients(roomName string) *Clients {
	return &Clients{
		clients:  make(map[*Client]bool),
		roomName: roomName,
	}
}

// GetClients returns all the clients in the room.
func (cs *Clients) GetClients() map[*Client]bool {
	return cs.clients
}

// Add adds a client to the room.
func (cs *Clients) Add(c *Client) {
	cs.clients[c] = true
}

// Remove removes a client from the room.
func (cs *Clients) Remove(c *Client) {
	if ok := cs.clients[c]; ok {
		delete(cs.clients, c)
		close(c.send)
	}
}

// NumClients returns the number of clients in the room.
func (cs *Clients) NumClients() int {
	return len(cs.clients)
}
