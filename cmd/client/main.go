package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
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
	gameState := gamelogic.NewGameState(username)
	gamelogic.PrintClientHelp()
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("can't subscribe to the queue: %v", err)
	}
	for {
		fmt.Println("Please enter the game command:")
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Error spawning unit: %v\n", err)
			}

		case "move":
			_, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("Move failed: %v\n", err)
			} else {
				fmt.Println("Move successful!")
			}

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Printf("Unknown command: '%s'. Type 'help' for a list of commands.\n", words[0])
		}
	}

	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, os.Interrupt)

	// fmt.Println("waiting for exit signal(ctrl+c)...")
	// sig := <-sigChan
	// fmt.Printf("\nsignal received: %v.Closing the program...\n", sig)
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
