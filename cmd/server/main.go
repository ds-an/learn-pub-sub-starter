package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	fmt.Println("Console on http://localhost:15672/")
	const connectionString = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	ch, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	gamelogic.PrintServerHelp()
	for {
		line := gamelogic.GetInput()
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case routing.PauseKey:
			log.Printf("Sending \"%s\" message", routing.PauseKey)
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Fatal(err)
			}
		case "resume":
			log.Print("Sending \"resume\" message")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Fatal(err)
			}
		case "quit":
			log.Print("Exiting the game")
			return
		default:
			log.Print("No such command")
		}
	}

	// fmt.Println("AMQP connection successful")
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("Shutting down Peril server...")
}
