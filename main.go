// +build darwin linux

// Hacked up version of
// https://godoc.org/golang.org/x/mobile/example/basic

package main

import (
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"log"
)

var (
	touchX float32
	touchY float32
	beginX float32
	beginY float32

	iHaveTheCard bool
)

func main() {
	app.Main(func(a app.App) {

		// managerImpl = game.NewV23Manager()
		screen := screen.NewScreen()
		// table := NewTable(managerImpl)

		// The server initializes with player '0' holding the card.
		iHaveTheCard = true

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
				}
			case paint.Event:
				screen.Paint(c, iHaveTheCard, touchX, touchY)
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
					beginX = e.X
					beginY = e.Y
					if e.X < 10 && e.Y < 10 {
						log.Printf("Shutting Down!\n")
						screen.Stop()
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
					log.Printf("  begin = (%v, %v)\n", beginX, beginY)
					log.Printf("    end = (%v, %v)\n", e.X, e.Y)
					log.Printf("  delta = (%v, %v)\n", e.X-beginX, e.Y-beginY)
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
				touchX = float32(c.WidthPx / 2)
				touchY = float32(c.HeightPx / 2)
				// gm.SetOrigin(touchX, touchY)
			}
		}
	})
}
