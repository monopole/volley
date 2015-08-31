// +build darwin linux

package main

import (
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/interpreter"
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"log"
)

func main() {
	app.Main(func(a app.App) {
		nsRoot := game.DetermineNamespaceRoot()
		log.Printf("Using v23.namespace.root=%s", nsRoot)
		interpreter.NewInterpreter(
			config.Chatty,
			game.NewV23Manager(
				config.Chatty, config.RootName, nsRoot),
			screen.NewScreen(),
		).Run(a)
	})
}
