package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	err = pubsub.PublishJSON(
		rabbitmqChannel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: true},
	)
	if err != nil {
		log.Fatalf("Couldn't publish JSON: %v", err)
	}

	log.Println("Starting Peril server...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Received signal %v. Exiting.\n", sig.String())
	os.Exit(0)

}
