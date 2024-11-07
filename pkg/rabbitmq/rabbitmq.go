package rabbitmq

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"image-compressions/internal/config"
)

func connectionRabbit(cfg *config.Configurations) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.Rabbitmq.Username, cfg.Rabbitmq.Password, cfg.Rabbitmq.Host, cfg.Rabbitmq.Port)
	conn, err := amqp.Dial(url)
	return conn, err
}

func setupChannel(ctx context.Context, conn *amqp.Connection, topic string) (*amqp.Channel, <-chan amqp.Delivery, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	msgs, err := ch.ConsumeWithContext(ctx, topic, "", true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	return ch, msgs, nil
}

func Consumer(ctx context.Context, cfg *config.Configurations) (<-chan amqp.Delivery, *amqp.Connection, error) {
	conn, err := connectionRabbit(cfg)
	if err != nil {
		return nil, nil, err
	}

	_, msgs, err := setupChannel(ctx, conn, cfg.Rabbitmq.Topic)
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
