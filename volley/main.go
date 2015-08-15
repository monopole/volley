// +build darwin linux

// Hacked up version of
// https://godoc.org/golang.org/x/mobile/example/basic

package main

import (
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/interpreter"
	"github.com/monopole/croupier/screen"
	"github.com/monopole/croupier/table"
	"golang.org/x/mobile/app"
	"log"
	"time"
)

func main() {
	app.Main(func(a app.App) {

		gm := game.NewV23Manager(
			config.Chatty, config.RootName, config.NamespaceRoot)
		gm.Initialize()

		if config.Chatty {
			log.Println("Making the table.")
		}
		s := screen.NewScreen()
		table := table.NewTable(
			config.Chatty,
			gm.Me(),
			s,
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

		// Let the network fire up?  Use channel to signal instead.
		<-time.After(3 * time.Second)

		ub := interpreter.NewInterpreter(
			config.Chatty,
			table.ChQuit(),
		)

		ub.Run(a, s)
	})
}
