// +build darwin linux

package main

import (
	"github.com/monopole/volley/config"
	"github.com/monopole/volley/engine"
	"github.com/monopole/volley/net"
	"github.com/monopole/volley/screen"
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
				config.Chatty, config.RootName, false, nsRoot),
			screen.NewScreen(),
		).Run(a)
	})
}
