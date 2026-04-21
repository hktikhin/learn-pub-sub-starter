package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

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

	msg := routing.PlayingState{
		IsPaused: true,
	}
	err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, msg)
	if err != nil {
		log.Printf("publish failed: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	fmt.Println("waiting for exit signal(ctrl+c)...")
	sig := <-sigChan
	fmt.Printf("\nsignal received: %v.Closing the program...\n", sig)
}
