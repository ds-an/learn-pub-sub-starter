package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	marshaledVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body: marshaledVal,
	})
	return err 
}

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType,
    handler func(T),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	deliveryCh, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
		)
	if err != nil {
		return err
	}
	go func() {
		for m := range deliveryCh {
			var body T	
			err := json.Unmarshal(m.Body, &body)
			if err != nil {
				log.Print(err)
				continue
			}
			handler(body)
			err = m.Ack(false)
			if err != nil {
				log.Print(err)
			}
		}
	}()
	return nil
}
