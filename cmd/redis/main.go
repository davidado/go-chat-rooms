// Description: Main entry point for the chat server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davidado/go-chat-rooms/config"
	"github.com/davidado/go-chat-rooms/pkg/api"
	"github.com/davidado/go-chat-rooms/pkg/auth"
	"github.com/davidado/go-chat-rooms/pkg/pubsub"
	"github.com/davidado/go-chat-rooms/service/chat"

	"github.com/redis/go-redis/v9"
)

func initRedis(ctx context.Context, conn *redis.Client) {
	if err := conn.Ping(ctx).Err(); err != nil {
		log.Fatal("unable to ping Redis:", err)
	}
	fmt.Println("successfully connected to Redis")
}

func main() {
	ctx := context.Background()
	conn := redis.NewClient(&redis.Options{
		Addr:     config.Envs.RedisHost,
		Password: config.Envs.RedisPassword, // no password set
		DB:       0,                         // use default DB
	})
	initRedis(ctx, conn)
	defer conn.Close()

	pubsub := pubsub.NewRedisPubSub(conn)
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
