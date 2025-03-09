package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitmqConnection = "amqp://guest:guest@localhost:5672/"

func main() {
	connection, err := amqp.Dial(rabbitmqConnection)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer connection.Close()
	log.Println("Conncetion to RabbitMQ is successful")

	log.Println("Starting Peril server...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Received signal %v. Exiting.\n", sig.String())
	os.Exit(0)

}
