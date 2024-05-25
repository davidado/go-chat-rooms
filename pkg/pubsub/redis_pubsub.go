package pubsub

import (
	"context"
	"fmt"
	"log"

	"github.com/davidado/go-chat-rooms/service/chat"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub is a Redis client for publishing and subscribing messages.
type RedisPubSub struct {
	ctx  context.Context
	conn *redis.Client
}

// NewRedisPubSub creates a new RedisPubSub client.
func NewRedisPubSub(ctx context.Context, conn *redis.Client) *RedisPubSub {
	return &RedisPubSub{
		ctx:  ctx,
		conn: conn,
	}
}

// Publish publishes a message to a topic.
func (ps *RedisPubSub) Publish(topic string, msg []byte) error {
	err := ps.conn.Publish(ps.ctx, topic, msg).Err()
	if err != nil {
		return err
	}
	return nil
}

// SubscribeAndBroadcast listens to a room (topic) for incoming messages
// then broadcasts them to the room. SubscribeAndBroadcast should only
// run once per room.
func (ps *RedisPubSub) SubscribeAndBroadcast(room *chat.Room) {
	subscriber := ps.conn.Subscribe(ps.ctx, room.Name)

	defer func() {
		err := subscriber.Unsubscribe(ps.ctx, room.Name)
		if err != nil {
			log.Println("SubscribeAndBroadcast - error cancelling consumer:", err)
		}
		fmt.Println("unsubscribed from topic", room.Name)
	}()

	fmt.Printf(" [*] %s waiting for messages.\n", room.Name)

	for {
		select {
		case <-room.Done:
			return
		case b := <-subscriber.Channel():
			msg := chat.Message{}
			err := msg.Unmarshal([]byte(b.Payload))
			if err != nil {
				log.Print("SubscribeAndBroadcast - error unmarshalling message:", err)
				continue
			}

			room.Broadcast(msg)
		}
	}
}
