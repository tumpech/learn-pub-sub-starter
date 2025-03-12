package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	chanDelivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer channel.Close()
		for delivery := range chanDelivery {
			var body T
			err := json.Unmarshal(delivery.Body, &body)
			if err != nil {
				fmt.Printf("could not unmarshall body: %v\n", err)
			}
			switch handler(body) {
			case Ack:
				delivery.Ack(false)
			case NackDiscard:
				delivery.Nack(false, false)
			case NackRequeue:
				delivery.Nack(false, true)
			}
		}
	}()

	return nil

}
