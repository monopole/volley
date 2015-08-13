// +build darwin linux

// Hacked up version of
// https://godoc.org/golang.org/x/mobile/example/basic

/*

The scheme will be

 1) Create some channels

 2) Create some objects, feeding them channels.

 3)   go obj1.run()
      go obj2.run()
      go obj3.run()

 4) Block on some quit channel.


Major objects are

*) Master (UserBuddy?) - converts app events into 'simpler' events

*) Screen - An object with a collection of methods
   for drawing.  No threads.  No channels.  Just a view.
   Access to this object only via the screen.

   Only one thread has a reference.

*) model.go   Player inteface

*) model.go:  Ball - just a simple object - has physics in it.
    not a thread.

*) model.go:  Game interface - Players, SendBall, Quit
   This is held by the table, and it protect the table
   from knowing about

*) V23Manager - Holds an instance of a v23 service, and
   many client instances.
   Uses an authorizer, dispatcher and initializer.
   Accepts and emits data on channels.
   Has a thread which....
   Implementation of a GameManager interface.  Why?

*) V23Player - holds N client interface to some other player
   Can send a ball to the other player.

*) GameBuddy - the v23 service defined in player.vdl,
   held by V23Manager
   Defines Recognize, Forget, and Accept.

*) server.go -- nameless implementation of GameBuddy
   impleements Recognize, Forget and Accept.
   Pulls stuff off the 'wire' and converts it into
   ptrs to objects with methods and sends them out
   on channels without blocking.
   An instance of this impl is passed to ifc.GameBuddeyServer


*) table.go  Thing that knows the positions of all the
             balls, which walls are transparent, where
             to send the balls, etc.

*/

package main

import (
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/interpreter"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"log"
)

const rootName = "volley/player"

// const namespaceRoot = "/104.197.96.113:3389"
// const namespaceRoot = "/172.17.166.64:23000"
const namespaceRoot = "/localhost:23000"

func main() {
	app.Main(func(a app.App) {

		gm := game.NewV23Manager(rootName, namespaceRoot)

		log.Println("Initializing game")

		chBall := make(chan *model.Ball)
		gm.Initialize(chBall)
		go gm.Run()

		chChQuit := gm.Quitter()

		screen := screen.NewScreen()
		// table := NewTable(managerImpl)

		interpreter := &interpreter.Interpreter{}
		go interpreter.doIt(chChQuit, a, screen)

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
