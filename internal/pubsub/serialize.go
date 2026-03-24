package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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
	unmarshaler := func(data []byte) (T, error) {
    var target T
    err := json.Unmarshal(data, &target)
    return target, err
	}
	return subscribe(conn, exchange, queueName, key, queueType, handler, unmarshaler)
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body: buf.Bytes(),
	})
	return err 
}

func SubscribeGob[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType,
    handler func(T) AckType,
) error {
	unmarshaler := func(data []byte) (T, error) {
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		var target T
		err := dec.Decode(&target)
		return target, err
	}
	return subscribe(conn, exchange, queueName, key, queueType, handler, unmarshaler)
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaler func([]byte) (T, error),
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
			body, err := unmarshaler(m.Body)
			if err != nil {
				log.Print(err)
				err = m.Nack(false, false)
				if err != nil {
					log.Print(err)
				} else {
					log.Print("Unmarshal error: message not acknowledged and discarded")
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
