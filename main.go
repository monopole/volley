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

		var c config.Event
		for e := range a.Events() {
			switch e := app.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					screen.Start()
				case lifecycle.CrossOff:
					screen.Stop()
				}
			case config.Event:
				c = e
				touchX = float32(c.WidthPx / 2)
				touchY = float32(c.HeightPx / 2)
				//gm.SetOrigin(touchX, touchY)
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
				touchX = e.X
				touchY = e.Y
				// }
			}

		}

	})
}
