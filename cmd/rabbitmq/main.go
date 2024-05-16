// Description: Main entry point for the chat server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davidado/go-chat-rooms/config"
	"github.com/davidado/go-chat-rooms/pkg/auth"
	"github.com/davidado/go-chat-rooms/pkg/pubsub"
	"github.com/davidado/go-chat-rooms/service/chat"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handleGetRoom(w http.ResponseWriter, r *http.Request) {
	dir, _ := os.Getwd()
	http.ServeFile(w, r, dir+"/public/chat.html")
}

func handleWebSocket(rooms *chat.Rooms, pubsub chat.PubSubber) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomName := r.PathValue("room")
		room, ok := rooms.Get(roomName)
		if !ok {
			room = chat.NewRoom(roomName, chat.NewClients(roomName), pubsub)
			rooms.Add(roomName, room)

			// SubscribeAndBroadcast should only run once per room.
			go room.Pubsub.SubscribeAndBroadcast(room)
		}
		go room.Run(rooms)
		chat.ServeWs(room, w, r)
	}
}

func main() {
	ctx := context.Background()
	conn, err := amqp.Dial(config.Envs.RabbitMQHost)
	if err != nil {
		log.Fatal("unable to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	pubsub := pubsub.NewRabbitMQPubSub(ctx, conn)
	rooms := chat.NewRooms()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{room}", handleGetRoom)
	mux.HandleFunc("GET /ws/{room}", auth.WithJWTAuth(handleWebSocket(rooms, pubsub)))

	srv := &http.Server{
		Addr:              config.Envs.Listen,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
