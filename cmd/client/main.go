package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const connectionString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	queueName := routing.PauseKey + "." + username
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.QueueTypeTransient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("AMQP connection successful")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Shutting down Peril client...")
}
