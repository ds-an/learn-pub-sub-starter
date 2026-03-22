package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	AckTypeAck AckType = iota
	AckTypeNackRequeue
	AckTypeNackDiscard
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
    handler func(T) AckType,
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
				err = m.Nack(false, false)
				if err != nil {
					log.Print(err)
				} else {
					log.Print("JSON unmarshal error: message not acknowledged and discarded")
				}
				continue
			}
			ackType := handler(body)
			switch ackType {
			case AckTypeAck:
				err = m.Ack(false)
				if err != nil {
					log.Print(err)
				} else {
					log.Print("Message acknowledged")
				}
			case AckTypeNackRequeue:
				err = m.Nack(false, true)
				if err != nil {
					log.Print(err)
				} else {
					log.Print("Message not acknowledged and requeued")
				}
			case AckTypeNackDiscard:
				err = m.Nack(false, false)
				if err != nil {
					log.Print(err)
				} else {
					log.Print("Message not acknowledged and discarded")
				}
			default:
				err = m.Nack(false, false)
				if err != nil {
					log.Print(err)
				} else {
					log.Print("Unexpexted AckType: message not acknowledged and discarded")
				}
			}
		}
	}()
	return nil
}
