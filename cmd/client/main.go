package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(move gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		case gamelogic.MoveOutcomeMakeWar:
			fmt.Println("Sending a war message because of the move...")
			warMsg := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			}
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername()), warMsg)
			if err != nil {
				log.Printf("publish failed: %v", err)
				return pubsub.NackRequeue

			}
			return pubsub.Ack
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			log.Printf("Unexpected move outcome: %v, discarding message", moveOutcome)
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(ch *amqp.Channel, username string, msg string) error {
	logEntry := routing.GameLog{
		CurrentTime: time.Now().UTC(),
		Message:     msg,
		Username:    username,
	}
	routingKey := fmt.Sprintf("%s.%s", routing.GameLogSlug, username)
	return pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routingKey,
		logEntry,
	)
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(warMsg gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(warMsg gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(warMsg)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			err := publishGameLog(ch, gs.GetUsername(), fmt.Sprintf("%s won a war against %s", winner, loser))
			if err != nil {
				fmt.Printf("Error publishing log: %v\n", err)
			}
			fmt.Printf("War outcome: Opponent won\n")
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			err := publishGameLog(ch, gs.GetUsername(), fmt.Sprintf("%s won a war against %s", winner, loser))
			if err != nil {
				fmt.Printf("Error publishing log: %v\n", err)
			}
			fmt.Printf("War outcome: You won!\n")
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			err := publishGameLog(ch, gs.GetUsername(), fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser))
			if err != nil {
				fmt.Printf("Error publishing log: %v\n", err)
			}
			fmt.Printf("War outcome: Draw\n")
			return pubsub.Ack

		default:
			log.Printf("Error: unexpected war outcome: %v", outcome)
			return pubsub.NackDiscard
		}
	}
}

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
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("cannot create channel: %v", err)
	}
	defer ch.Close()
	fmt.Println("mq channel created")
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
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.Transient,
		handlerMove(gameState, ch),
	)
	if err != nil {
		log.Fatalf("can't subscribe to the queue: %v", err)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s", routing.WarRecognitionsPrefix),
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.Durable,
		handlerWar(gameState, ch),
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
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("Move failed: %v\n", err)
			} else {
				fmt.Println("Move successful!")
			}
			fmt.Println("Sending a move message...")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), move)
			if err != nil {
				log.Printf("publish failed: %v", err)
			}

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(words) < 2 {
				fmt.Println("Usage: spam <number>")
				continue
			}

			n, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("Error: %v is not a valid number\n", words[1])
				continue
			}

			fmt.Printf("Spamming %d malicious logs...\n", n)

			for i := range n {
				message := gamelogic.GetMaliciousLog()

				err := publishGameLog(ch, gameState.GetUsername(), message)
				if err != nil {
					fmt.Printf("Error sending spam log %d: %v\n", i, err)
					break
				}
			}
			fmt.Println("Spamming complete.")

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
