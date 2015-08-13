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
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"log"
)

type UserBuddy struct {
	touchX float32
	touchY float32
	beginX float32
	beginY float32

	iHaveTheCard bool
}

func (ub *UserBuddy) quit(chChQuit chan chan bool) {
	chQuit := make(chan bool)
	chChQuit <- chQuit
	<-chQuit
}

func (ub *UserBuddy) doIt(
	chChQuit chan chan bool, a app.App, screen *screen.Screen) {
	// The server initializes with player '0' holding the card.
	ub.iHaveTheCard = true

	log.Printf("Hi there.\n")

	grabbingVector := false
	var c config.Event
	for e := range a.Events() {
		switch e := app.Filter(e).(type) {
		case lifecycle.Event:
			switch e.Crosses(lifecycle.StageVisible) {
			case lifecycle.CrossOn:
				log.Printf("Starting Up!\n")
				screen.Start()
			case lifecycle.CrossOff:
				log.Printf("Shutting Down!\n")
				screen.Stop()
				ub.quit(chChQuit)
				return
			}
		case paint.Event:
			screen.Paint(c, ub.iHaveTheCard, ub.touchX, ub.touchY)
			a.EndPaint(e)
		case touch.Event:
			// if e.Type == touch.TypeEnd && iHaveTheCard {
			// gm.PassTheCard()
			// touchX = gm.GetOriginX()
			// touchY = gm.GetOriginY()
			// iHaveTheCard = false
			// } else {
			// touchX = e.X
			// touchY = e.Y
			// }
			switch e.Type {
			case touch.TypeBegin:
				grabbingVector = true
				log.Printf("Begin.\n")
				ub.beginX = e.X
				ub.beginY = e.Y
				if e.X < 10 && e.Y < 10 {
					log.Printf("Shutting Down!\n")
					screen.Stop()
					ub.quit(chChQuit)
					return
				}
			case touch.TypeMove:
				log.Printf("Moving.\n")
			case touch.TypeEnd:
				if !grabbingVector {
					log.Printf("That's odd!\n")
				}
				grabbingVector = false
				log.Printf("Done\n")
				log.Printf("  begin = (%v, %v)\n", ub.beginX, ub.beginY)
				log.Printf("    end = (%v, %v)\n", e.X, e.Y)
				log.Printf("  delta = (%v, %v)\n", e.X-ub.beginX, e.Y-ub.beginY)
				/*
						On X11, screen points come in as some kind of pixels.
						As the screen is resized, 0,0 stays the same,
					  but the other numbers change.

							(0,0)    ... (high, 0)
							...          ...
							(0,high) ... (high, high)

				*/
			}
		case config.Event:
			c = e
			// These numbers are in the same units as touch events.
			// After a resize,
			//   e.X  <= c.WidthPx
			//   e.Y  <= c.HeightPx
			log.Printf(" config = (%v, %v)\n", c.WidthPx, c.HeightPx)
			ub.touchX = float32(c.WidthPx / 2)
			ub.touchY = float32(c.HeightPx / 2)
			// gm.SetOrigin(touchX, touchY)
		}
	}
}

func main() {
	app.Main(func(a app.App) {

		chChQuit := make(chan chan bool)
		chBall := make(chan *model.Ball)

		gm := game.NewV23Manager(chChQuit)
		gm.Initialize(chBall)
		screen := screen.NewScreen()
		// table := NewTable(managerImpl)

		ub := &UserBuddy{}
		go ub.doIt(chChQuit, a, screen)
		select {}

	})
}
