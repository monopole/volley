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
	//	"time"
)

const rootName = "croupier/player"

// const namespaceRoot = "/104.197.96.113:3389"
// const namespaceRoot = "/172.17.166.64:23000"
const namespaceRoot = "/localhost:23000"

func main() {

	chChQuit := make(chan chan bool)

	gm := game.NewV23Manager(rootName, namespaceRoot, chChQuit)

	log.Println("Initializing game")

	chBall := make(chan *model.Ball)
	gm.Initialize(chBall)
	select {}
	//	go gm.Run()
	//	for i := 10; i < 17; i++ {
	//		<-time.After(3 * time.Second)
	//		chBall <- model.NewBall(model.NewPlayer(i), 1, 2, 3, 4)
	//	}
}
