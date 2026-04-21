package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
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
	username, err := gamelogic.ClientWelcome()
}
