// File: pkg/rabbitmq/rabbitmq.go

package rabbitmq

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"image-compressions/internal/config"
)

func connectionRabbit(cfg *config.Configurations) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.Rabbitmq.Username,
		cfg.Rabbitmq.Password,
		cfg.Rabbitmq.Host,
		cfg.Rabbitmq.Port,
	)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed connection to RabbitMQ: %w", err)
	}
	return conn, nil
}

func StartConsumer(ctx context.Context, cfg *config.Configurations, poolSize int) (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {
	conn, err := connectionRabbit(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Setting Quality of Service (QoS)
	// this tells RabbitMQ to only send `poolSize` messages to this consumer
	// before waiting for acknowledgement. This prevents our workers from being flooded.
	err = ch.Qos(
		poolSize, // prefetchCount: Maximum number of messages sent
		0,        // prefetchSize: Total message size (0 = infinite)
		false,    // global: Does this setting apply to all consumers on the channel (false = only for this consumer)
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	// This is the key to manual acknowledgment.
	msgs, err := ch.ConsumeWithContext(ctx,
		cfg.Rabbitmq.Topic,
		"",    // consumer tag
		false, // autoAck: change from `true` to `false`
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, nil, fmt.Errorf("failed to set consume: %w", err)
	}

	return msgs, conn, ch, nil
}
