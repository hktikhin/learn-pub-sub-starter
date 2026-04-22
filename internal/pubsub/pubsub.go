package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func (sqt SimpleQueueType) String() string {
	switch sqt {
	case Durable:
		return "durable"
	case Transient:
		return "transient"
	default:
		return "unknown"
	}
}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}
	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		msg,
	)
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Error opening channel: %v", err)
		return nil, amqp.Queue{}, err
	}
	isDurable := (queueType == Durable)
	isAutoDelete := (queueType == Transient)
	isExclusive := (queueType == Transient)

	q, err := ch.QueueDeclare(
		queueName,    // name
		isDurable,    // durable
		isAutoDelete, // autoDelete
		isExclusive,  // exclusive
		false,        // noWait
		nil,          // args
	)
	if err != nil {
		log.Printf("Error declaring queue %s: %v", queueName, err)
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(
		q.Name,   // name
		key,      // key
		exchange, // exchange
		false,    // noWait
		nil,      // args
	)
	if err != nil {
		log.Printf("Error binding queue %s to exchange %s: %v", queueName, exchange, err)
		return nil, amqp.Queue{}, err
	}

	log.Printf("Successfully declared and bound queue: %s (Type: %v)", queueName, queueType)
	return ch, q, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	ch, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		log.Printf("can't declare and bind queue: %v", err)
		return err
	}
	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %v", err)
		ch.Close()
		return err
	}
	go func() {
		for d := range msgs {
			defer ch.Close()
			var msg T
			json.Unmarshal(d.Body, &msg)
			log.Printf("Received a message: %v", msg)
			handler(msg)
			err = d.Ack(false)
			if err != nil {
				log.Printf("Failed to ack the message: %v", err)
			}
		}
	}()
	return nil
}
