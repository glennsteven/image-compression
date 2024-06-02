package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"image-compressions/internal/config"
)

func connectionRabbit(cfg *config.Configurations) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.Rabbitmq.Username, cfg.Rabbitmq.Password, cfg.Rabbitmq.Host, cfg.Rabbitmq.Port)
	conn, err := amqp.Dial(url)
	return conn, err
}

func setupChannel(conn *amqp.Connection, topic string) (*amqp.Channel, <-chan amqp.Delivery, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	msgs, err := ch.Consume(topic, "", true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	return ch, msgs, nil
}

func Consumer(cfg *config.Configurations) (<-chan amqp.Delivery, *amqp.Connection, error) {
	conn, err := connectionRabbit(cfg)
	if err != nil {
		return nil, nil, err
	}

	_, msgs, err := setupChannel(conn, cfg.Rabbitmq.Topic)
	if err != nil {
		return nil, nil, err
	}

	bidirectionalMsgs := make(chan amqp.Delivery)
	go func() {
		defer close(bidirectionalMsgs)
		for msg := range msgs {
			bidirectionalMsgs <- msg
		}
	}()

	return bidirectionalMsgs, conn, nil
}
