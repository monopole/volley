package main

import (
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/table"
	"log"
	"time"
)

func main() {
	gm := game.NewV23Manager(
		config.Chatty, config.RootName, config.NamespaceRoot)
	gm.Initialize()

	if config.Chatty {
		log.Println("Making the table.")
	}
	table := table.NewTable(
		config.Chatty,
		gm.Me(),
		nil, // screen.NewScreen(),
		gm.ChIncomingBall(),
		gm.ChDoorCommand(),
		gm.ChQuit(),
	)

	if config.Chatty {
		log.Println("Firing table")
	}
	go table.Run()

	if config.Chatty {
		log.Println("Firing v23")
	}
	go gm.Run(table.ChBallCommand())

	delta := 5
	timeStep := time.Duration(delta) * time.Second
	for i := 4; i > 0; i-- {
		if config.Chatty {
			log.Printf("%d seconds left...\n", i*delta)
		}
		<-time.After(timeStep)
	}

	if config.Chatty {
		log.Println("Sending quit.")
	}
	ch := make(chan bool)
	table.ChQuit() <- ch
	<-ch

	if config.Chatty {
		log.Println("All done.")
	}
}
