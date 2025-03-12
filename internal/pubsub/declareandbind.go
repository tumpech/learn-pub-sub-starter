package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	rabbitmqChannel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	var rabbitmqQueue amqp.Queue
	table := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	if simpleQueueType == SimpleQueueDurable {
		rabbitmqQueue, err = rabbitmqChannel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			table,
		)
	} else {
		rabbitmqQueue, err = rabbitmqChannel.QueueDeclare(
			queueName,
			false,
			true,
			true,
			false,
			table,
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
