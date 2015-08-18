// +build darwin linux

// Hacked up version of
// https://godoc.org/golang.org/x/mobile/example/basic
// https://github.com/golang/mobile/tree/master/example

package main

import (
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/interpreter"
	"github.com/monopole/croupier/screen"
	"github.com/monopole/croupier/table"
	"golang.org/x/mobile/app"
	"log"
	"net/http"
)

func gotNetwork() bool {
	_, err := http.Get(config.TestDomain)
	if err == nil {
		log.Printf("Network up - able to hit %s", config.TestDomain)
		return true
	}
	log.Printf("Something wrong with network: %v", err)
	return false
}

func main() {
	if config.FailFast {
		if !gotNetwork() {
			return
		}
	}
	app.Main(func(a app.App) {

		// All v23 dependence here.
		gm := game.NewV23Manager(
			config.Chatty, config.RootName, config.NamespaceRoot)

		// Calls v23.Init(), determines current players from MT, etc.
		gm.Initialize()

		// No GL, mobile or v23.  Contains physics, notion of multiple
		// screens, etc.  Handles applying impulse to ball, sending
		// it to another player, etc.
		table := table.NewTable(
			config.Chatty,
			gm.Me(),
			// All GL dependence in screen (mockable).
			screen.NewScreen(),
			gm.ChIncomingBall(),
			gm.ChDoorCommand(),
			gm.ChQuit(),
		)

		go table.Run()
		go gm.Run(table.ChBallCommand())

		// Event loop - converts KM events to table commands.
		interpreter.NewInterpreter(
			config.Chatty,
			table.ChQuit(),
			table.ChExecCommand(),
			table.ChImpulse(),
			table.ChResize(),
		).Run(a)
	})
}
