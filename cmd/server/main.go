package main

import (
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
	rabbitmqConnection, err := amqp.Dial(rabbitmqConnectionString)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer rabbitmqConnection.Close()
	log.Println("Conncetion to RabbitMQ is successful")
	rabbitmqChannel, err := rabbitmqConnection.Channel()
	if err != nil {
		log.Fatalf("Error creating channel in RabbitMQ: %v", err)
	}

	log.Println("Starting Peril server...")

	_, queue, err := pubsub.DeclareAndBind(
		rabbitmqConnection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.DurableQueue,
	)

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	log.Printf("Queue %v declared and bound!\n", queue.Name)

	gamelogic.PrintServerHelp()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v. Exiting.\n", sig.String())
		os.Exit(0)
	}()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			log.Print("Pausing game...")
			err = pubsub.PublishJSON(
				rabbitmqChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf("Couldn't publish JSON: %v", err)
			}
			continue
		case "resume":
			log.Print("Resuming game...")
			err = pubsub.PublishJSON(
				rabbitmqChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Fatalf("Couldn't publish JSON: %v", err)
			}
			continue
		case "quit":
			log.Println("Exiting...")
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
