package main

// $VEGGIE/bin/mounttabled --v23.tcp.address :23000 &
//
// $VEGGIE/bin/namespace --v23.namespace.root '/localhost:23000' glob -l '*/*'
//
// $VEGGIE/bin/beans

import (
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/model"
	"log"
	"time"
)

const rootName = "croupier/player"

// const namespaceRoot = "/104.197.96.113:3389"
// const namespaceRoot = "/172.17.166.64:23000"
const namespaceRoot = "/localhost:23000"

func main() {

	gm := game.NewV23Manager(rootName, namespaceRoot)

	log.Println("Initializing game")

	chBall := make(chan *model.Ball)
	gm.Initialize(chBall)
	go gm.Run()

	delta := 5
	timeStep := time.Duration(delta) * time.Second
	for i := 6; i > 0; i-- {
		log.Printf("%d seconds left...\n", i*delta)
		<-time.After(timeStep)
	}

	log.Println("Sending quit.")
	ch := make(chan bool)
	gm.Quitter() <- ch
	<-ch

	log.Println("All done.")
}
