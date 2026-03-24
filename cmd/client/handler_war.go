package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(dw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.AckTypeNackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.AckTypeNackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			logMessage := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := publishGameLog(ch, gs.GetUsername(), logMessage)
			if err != nil {
				return pubsub.AckTypeNackRequeue
			}
			return pubsub.AckTypeAck
		case gamelogic.WarOutcomeYouWon:
			logMessage := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := publishGameLog(ch, gs.GetUsername(), logMessage)
			if err != nil {
				return pubsub.AckTypeNackRequeue
			}
			return pubsub.AckTypeAck
		case gamelogic.WarOutcomeDraw:
			logMessage := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := publishGameLog(ch, gs.GetUsername(), logMessage)
			if err != nil {
				return pubsub.AckTypeNackRequeue
			}
			return pubsub.AckTypeAck
		default:
			log.Print("Error: war outcome type not recognized")
			return pubsub.AckTypeNackDiscard
		}
	}
}

func publishGameLog(ch *amqp.Channel, username, message string) error {
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message: message,
		Username: username,
	}
	return pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		log,
	)
}
