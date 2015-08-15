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
)

func main() {
	app.Main(func(a app.App) {
		gm := game.NewV23Manager(
			config.Chatty, config.RootName, config.NamespaceRoot)

		gm.Initialize() // Calls v23.Init()

		table := table.NewTable(
			config.Chatty,
			gm.Me(),
			screen.NewScreen(), // All the GL dependence in here.
			gm.ChIncomingBall(),
			gm.ChDoorCommand(),
			gm.ChQuit(),
		)

		go table.Run()
		go gm.Run(table.ChBallCommand())

		interpreter.NewInterpreter(
			config.Chatty,
			table.ChQuit(),
			table.ChExecCommand(),
			table.ChImpulse(),
			table.ChResize(),
		).Run(a)
	})
}
