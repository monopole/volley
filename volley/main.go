// +build darwin linux

package main

import (
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/engine"
	"github.com/monopole/croupier/net"
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"log"
)

func main() {
	app.Main(func(a app.App) {
		nsRoot := "/" + net.DetermineNamespaceRoot()
		log.Printf("Using v23.namespace.root=%s", nsRoot)
		engine.NewEngine(
			config.Chatty,
			net.NewV23Manager(
				config.Chatty, config.RootName, nsRoot),
			screen.NewScreen(),
		).Run(a)
	})
}
