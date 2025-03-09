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

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)

	pubsub.DeclareAndBind(
		rabbitmqConnection,
		"peril_direct",
		queueName,
		routing.PauseKey,
		pubsub.TransientQueue,
	)

	fmt.Println("Starting Peril client...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Received signal %v. Exiting.\n", sig.String())
	os.Exit(0)
}
