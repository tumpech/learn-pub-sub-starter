package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DurableQueue   = iota // 0
	TransientQueue        // 1
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	marshalledT, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        marshalledT,
		},
	)
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	rabbitmqChannel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	var rabbitmqQueue amqp.Queue

	if simpleQueueType == DurableQueue {
		rabbitmqQueue, err = rabbitmqChannel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
	} else {
		rabbitmqQueue, err = rabbitmqChannel.QueueDeclare(
			queueName,
			false,
			true,
			true,
			false,
			nil,
		)
	}

	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = rabbitmqChannel.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)

	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return rabbitmqChannel, rabbitmqQueue, nil

}
