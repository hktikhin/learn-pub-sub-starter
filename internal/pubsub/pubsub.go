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

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
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
	handler func(T) AckType,
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
		defer ch.Close()
		for d := range msgs {
			var msg T
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("JSON unmarshal error: %v", err)
				d.Nack(false, false)
				continue
			}
			log.Printf("Received a message: %v", msg)
			action := handler(msg)
			var ackErr error
			switch action {
			case Ack:
				log.Printf("Acking message")
				ackErr = d.Ack(false)
			case NackRequeue:
				log.Printf("Nacking message (requeueing)")
				ackErr = d.Nack(false, true)
			case NackDiscard:
				log.Printf("Nacking message (discarding)")
				ackErr = d.Nack(false, false)
			default:
				log.Printf("Unknown acktype: %v, defaulting to NackRequeue", action)
				ackErr = d.Nack(false, true)
			}
			if ackErr != nil {
				log.Printf("Failed to ack the message: %v", ackErr)
			}
		}
	}()
	return nil
}
