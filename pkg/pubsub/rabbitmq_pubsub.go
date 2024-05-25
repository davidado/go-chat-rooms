// Package pubsub provides a publish/subscribe interface for sending messages.
package pubsub

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidado/go-chat-rooms/service/chat"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPubSub is a Redis client for publishing and subscribing messages.
type RabbitMQPubSub struct {
	ctx  context.Context
	conn *amqp.Connection
}

// NewRabbitMQPubSub creates a new RabbitMQPubSub client.
func NewRabbitMQPubSub(ctx context.Context, conn *amqp.Connection) *RabbitMQPubSub {
	return &RabbitMQPubSub{
		ctx:  ctx,
		conn: conn,
	}
}

func configureQueue(ch *amqp.Channel, topic string) (*amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		topic, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Publish publishes a message to a topic.
func (ps *RabbitMQPubSub) Publish(topic string, msg []byte) error {
	ch, err := ps.conn.Channel()
	if err != nil {
		log.Println("error creating channel:", err)
		return err
	}
	defer ch.Close()

	q, err := configureQueue(ch, topic)
	if err != nil {
		log.Println("error declaring queue:", err)
		return err
	}

	ctx, cancel := context.WithTimeout(ps.ctx, 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		"", // exchange
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	if err != nil {
		log.Println("error publishing message:", err)
		return err
	}
	return nil
}

// SubscribeAndBroadcast listens to a room (topic) for incoming messages
// then broadcasts them to the room. SubscribeAndBroadcast should only
// run once per room.
func (ps *RabbitMQPubSub) SubscribeAndBroadcast(room *chat.Room) {
	ch, err := ps.conn.Channel()
	if err != nil {
		log.Println("error creating channel:", err)
		return
	}

	defer func() {
		err := ch.Cancel(room.Name, false)
		if err != nil {
			log.Println("error cancelling consumer:", err)
		}
		ch.Close()
		fmt.Println("unsubscribed from topic", room.Name)
	}()

	q, err := configureQueue(ch, room.Name)
	if err != nil {
		log.Println("error declaring queue:", err)
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Println("error consuming messages:", err)
		return
	}

	fmt.Printf(" [*] %s waiting for messages.\n", room.Name)

	for {
		select {
		case <-room.Done:
			return
		case d := <-msgs:
			msg := chat.Message{}
			err = msg.Unmarshal(d.Body)
			if err != nil {
				log.Print("error unmarshalling message:", err)
				continue
			}

			room.Broadcast(msg)
		}
	}
}
