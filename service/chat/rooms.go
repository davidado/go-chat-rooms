package chat

// Rooms maintains the set of active rooms.
type Rooms struct {
	// Rooms.
	rooms map[string]*Room
}

// NewRooms creates a new rooms store.
func NewRooms() *Rooms {
	return &Rooms{
		rooms: make(map[string]*Room),
	}
}

// GetAll returns all the rooms.
func (rs *Rooms) GetAll() map[string]*Room {
	return rs.rooms
}

// Get returns a room based on the room name.
func (rs *Rooms) Get(name string) (*Room, bool) {
	room, ok := rs.rooms[name]
	return room, ok
}

// Add adds a room to the room store.
func (rs *Rooms) Add(name string, room *Room) {
	rs.rooms[name] = room
}

// Remove removes a room from the room store.
func (rs *Rooms) Remove(name string) {
	delete(rs.rooms, name)
}
