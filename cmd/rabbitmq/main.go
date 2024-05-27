// Description: Main entry point for the chat server.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/davidado/go-chat-rooms/config"
	"github.com/davidado/go-chat-rooms/pkg/api"
	"github.com/davidado/go-chat-rooms/pkg/auth"
	"github.com/davidado/go-chat-rooms/pkg/pubsub"
	"github.com/davidado/go-chat-rooms/service/chat"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial(config.Envs.RabbitMQHost)
	if err != nil {
		log.Fatal("unable to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	pubsub := pubsub.NewRabbitMQPubSub(conn)
	rooms := chat.NewRooms()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{room}", api.HandleGetRoom)
	mux.HandleFunc("GET /ws/{room}", auth.WithJWTAuth(api.HandleWebSocket(rooms, pubsub)))

	srv := &http.Server{
		Addr:              config.Envs.Listen,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
