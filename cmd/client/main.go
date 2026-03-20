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
	fmt.Println("Starting Peril client...")
	const connectionString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(username)
	queueName := routing.PauseKey + "." + username
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.QueueTypeTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatal(err)
	}

	queueNameArmyMoves := routing.ArmyMovesPrefix + "." + username
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		queueNameArmyMoves,
		routing.ArmyMovesPrefix+".*",
		pubsub.QueueTypeTransient,
		handlerMove(gameState),
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				log.Print(err)
			}
		case "move":
			armyMove, err := gameState.CommandMove(words)
			if err != nil {
				log.Print(err)
				continue
			}
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				armyMove,
			)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Move %v successfully executed", armyMove)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Print("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Print("No such command")
		}
	}

	// fmt.Println("AMQP connection successful")
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("Shutting down Peril client...")
}
