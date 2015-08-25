// +build darwin linux

package main

import (
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/interpreter"
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
)

func main() {
	app.Main(func(a app.App) {
		interpreter.NewInterpreter(
			config.Chatty,
			game.NewV23Manager(
				config.Chatty, config.RootName, config.NamespaceRoot),
			screen.NewScreen(),
		).Run(a)
	})
}
