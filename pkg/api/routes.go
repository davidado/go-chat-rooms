// Package api provides the API handlers for the chat service.
package api

import (
	"context"
	"net/http"
	"os"

	"github.com/davidado/go-chat-rooms/service/chat"
)

// HandleGetRoom serves the chat room page.
func HandleGetRoom(w http.ResponseWriter, r *http.Request) {
	dir, _ := os.Getwd()
	http.ServeFile(w, r, dir+"/public/chat.html")
}

// HandleWebSocket handles websocket connections.
func HandleWebSocket(rooms *chat.Rooms, pubsub chat.PubSubber) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomName := r.PathValue("room")
		room, ok := rooms.Get(roomName)
		if !ok {
			ctx, cancel := context.WithCancel(context.Background())
			room = chat.NewRoom(ctx, cancel, roomName, pubsub)
			room.Subscribe()
			rooms.Add(roomName, room)
		}

		// Each client needs its own room.Run instance
		// but there should only ever be one room.
		go room.Run(rooms)
		chat.ServeWs(room, w, r)
	}
}
