package pkg

import (
	"fmt"
	"github.com/streadway/amqp"
)

func connectionRabbit(config RabbitMQConfig) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Username, config.Password, config.Host, config.Port)
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

func Consumer(config RabbitMQConfig, topic string) (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {
	conn, err := connectionRabbit(config)
	if err != nil {
		return nil, nil, nil, err
	}

	ch, msgs, err := setupChannel(conn, topic)
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}

	bidirectionalMsgs := make(chan amqp.Delivery)
	go func() {
		defer close(bidirectionalMsgs)
		for msg := range msgs {
			bidirectionalMsgs <- msg
		}
	}()

	return bidirectionalMsgs, conn, ch, nil
}

type RabbitMQConfig struct {
	Username string
	Password string
	Port     string
	Host     string
}

func NewRabbitMQConfig(username, password, port, host string) RabbitMQConfig {
	return RabbitMQConfig{
		Username: username,
		Password: password,
		Port:     port,
		Host:     host,
	}
}
