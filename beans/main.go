package main

// $VEGGIE/bin/mounttabled --v23.tcp.address :23000 &
//
// $VEGGIE/bin/namespace --v23.namespace.root '/localhost:23000' glob -l '*/*'
//
// $VEGGIE/bin/beans

import (
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/table"
	"log"
	"time"
)

const rootName = "beans/player"

// const namespaceRoot = "/104.197.96.113:3389"
// const namespaceRoot = "/172.17.166.64:23000"
// const namespaceRoot = "/192.168.2.71:23000"
// const namespaceRoot = "/localhost:23000"
const namespaceRoot = "/192.168.2.71:23000"

func main() {
	log.Println("Making v23.")
	gm := game.NewV23Manager(rootName, namespaceRoot)
	gm.Initialize()

	log.Println("Making the table.")
	table := table.NewTable(
		gm.Me(),
		nil, // screen.NewScreen(),
		gm.ChIncomingBall(),
		gm.ChDoorCommand(),
		gm.ChQuit(),
	)

	log.Println("Firing table")
	go table.Run()

	log.Println("Firing v23")
	go gm.Run(table.ChBallCommand())

	delta := 5
	timeStep := time.Duration(delta) * time.Second
	for i := 4; i > 0; i-- {
		log.Printf("%d seconds left...\n", i*delta)
		<-time.After(timeStep)
	}

	log.Println("Sending quit.")
	ch := make(chan bool)
	table.ChQuit() <- ch
	<-ch

	log.Println("All done.")
}
