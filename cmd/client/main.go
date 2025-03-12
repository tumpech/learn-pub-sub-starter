package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitmqConnectionString = "amqp://guest:guest@localhost:5672/"

func main() {
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error prompting user for username: %v", err)
	}

	rabbitmqConnection, err := amqp.Dial(rabbitmqConnectionString)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}

	rabbitmqChannel, err := rabbitmqConnection.Channel()
	if err != nil {
		log.Fatalf("Error creating channel in RabbitMQ: %v", err)
	}

	fmt.Println("Starting Peril client...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v. Exiting.\n", sig.String())
		os.Exit(0)
	}()

	gamestate := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		rabbitmqConnection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gamestate),
	)

	if err != nil {
		log.Fatalf("Error subscribing to queue: %v", err)
	}

	err = pubsub.SubscribeJSON(
		rabbitmqConnection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.SimpleQueueTransient,
		handlerMove(rabbitmqChannel, gamestate),
	)

	if err != nil {
		log.Fatalf("Error subscribing to queue: %v", err)
	}

	err = pubsub.SubscribeJSON(
		rabbitmqConnection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.SimpleQueueDurable,
		handlerWar(rabbitmqChannel, gamestate),
	)

	if err != nil {
		log.Fatalf("Error subscribing to queue: %v", err)
	}

	gamelogic.PrintClientHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			err = gamestate.CommandSpawn(input)
			if err != nil {
				log.Printf("Error while executing command: %v", err)
				continue
			}
			continue
		case "move":
			move, err := gamestate.CommandMove(input)
			if err != nil {
				log.Printf("Error while executing command: %v", err)
				continue
			}
			err = pubsub.PublishJSON(
				rabbitmqChannel,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				move,
			)
			if err != nil {
				log.Fatalf("Couldn't publish JSON: %v", err)
			}
			log.Printf("Moved successfully.")
			continue
		case "status":
			gamestate.CommandStatus()
			continue
		case "spam":
			log.Printf("Spamming not allowed yet!")
			continue
		case "quit":
			gamelogic.PrintQuit()
			os.Exit(0)
		case "help":
			gamelogic.PrintServerHelp()
			continue
		default:
			log.Printf("Unknown command: %s", input[0])
			continue
		}
	}

}
