package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerLog() func(gl routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			fmt.Printf("Error writing log to disk: %v\n", err)

			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}

func main() {
	fmt.Println("Starting Peril server...")

	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("can't connect to RabbitMQ: %v", err)
	}
	defer func() {
		conn.Close()
		fmt.Println("RabbitMQ connection closed")
	}()
	fmt.Println("connected to RabbitMQ")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("cannot create channel: %v", err)
	}
	defer ch.Close()
	fmt.Println("mq channel created")
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		handlerLog(),
	)
	if err != nil {
		log.Fatalf("could not subscribe to game logs: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		// fmt.Println("Please enter the server command:")
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Sending a pause message...")
			msg := routing.PlayingState{
				IsPaused: true,
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, msg)
			if err != nil {
				log.Printf("publish failed: %v", err)
			}

		case "resume":
			fmt.Println("Sending a resume message...")
			msg := routing.PlayingState{
				IsPaused: false,
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, msg)
			if err != nil {
				log.Printf("publish failed: %v", err)
			}

		case "quit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Printf("I don't understand the command: %s\n", words[0])
		}
	}

	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, os.Interrupt)

	// fmt.Println("waiting for exit signal(ctrl+c)...")
	// sig := <-sigChan
	// fmt.Printf("\nsignal received: %v.Closing the program...\n", sig)
}
