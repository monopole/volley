// +build darwin linux

// Hacked up version of
// https://godoc.org/golang.org/x/mobile/example/basic

package main

import (
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/interpreter"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"log"
	"time"
)

const rootName = "volley/player"

// const namespaceRoot = "/104.197.96.113:3389"
// const namespaceRoot = "/172.17.166.64:23000"
const namespaceRoot = "/localhost:23000"

func main() {
	app.Main(func(a app.App) {

		log.Println("Starting.")

		gm := game.NewV23Manager(rootName, namespaceRoot)
		gm.Initialize()
		go gm.Run()
		chChQuit := gm.Quitter()

		table := table.NewTable(
			gm.Me(),
			screen.NewScreen(),
			nil, nil, nil, nil,
		)

		//		interpreter := &interpreter.Interpreter{}
		//	go interpreter.doIt(chChQuit, a, screen)

		delta := 5
		timeStep := time.Duration(delta) * time.Second
		for i := 6; i > 0; i-- {
			log.Printf("%d seconds left...\n", i*delta)
			<-time.After(timeStep)
		}

		log.Println("Sending quit.")
		ch := make(chan bool)
		chChQuit <- ch
		<-ch
		log.Println("All done.")

	})
}
