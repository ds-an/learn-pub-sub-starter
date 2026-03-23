package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)

func handlerWar(gs *gamelogic.GameState) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(dw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.AckTypeNackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.AckTypeNackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.AckTypeAck
		case gamelogic.WarOutcomeYouWon:
			return pubsub.AckTypeAck
		case gamelogic.WarOutcomeDraw:
			return pubsub.AckTypeAck
		default:
			log.Print("Error: war outcome type not recognized")
			return pubsub.AckTypeNackDiscard
		}
	}
}
