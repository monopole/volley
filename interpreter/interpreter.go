package interpreter

import (
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"log"
)

type Interpreter struct {
	chatty bool
	chQuit chan<- chan bool // Not owned, written to.
	touchX float32
	touchY float32
	beginX float32
	beginY float32
}

func NewInterpreter(
	chatty bool,
	chQuit chan<- chan bool,
) *Interpreter {
	return &Interpreter{
		chatty,
		chQuit,
		0, 0, 0, 0,
	}
}

func (ub *Interpreter) quit() {
	ch := make(chan bool)
	ub.chQuit <- ch
	<-ch
}

func (ub *Interpreter) Run(
	a app.App,
	screen *screen.Screen) {

	if ub.chatty {
		log.Printf("Hi there.\n")
	}
	grabbingVector := false
	var sz size.Event
	for e := range a.Events() {
		switch e := app.Filter(e).(type) {
		case lifecycle.Event:
			switch e.Crosses(lifecycle.StageVisible) {
			case lifecycle.CrossOn:
				if ub.chatty {
					log.Printf("App starting!\n")
				}
				screen.Start()
			case lifecycle.CrossOff:
				if ub.chatty {
					log.Printf("App stopping!\n")
				}
				screen.Stop()
				ub.quit()
				return
			}
		case paint.Event:
			// screen.Paint(sz, ub.touchX, ub.touchY)
			a.EndPaint(e)
		case touch.Event:
			// if e.Type == touch.TypeEnd {
			// gm.PassTheCard()
			// touchX = gm.GetOriginX()
			// touchY = gm.GetOriginY()
			// } else {
			// touchX = e.X
			// touchY = e.Y
			// }
			switch e.Type {
			case touch.TypeBegin:
				grabbingVector = true
				if ub.chatty {
					log.Printf("Touch Begin.\n")
				}
				ub.beginX = e.X
				ub.beginY = e.Y
				if e.X < 10 && e.Y < 10 {
					if ub.chatty {
						log.Printf("Touched shutdown spot.\n")
					}
					screen.Stop()
					ub.quit()
					return
				}
			case touch.TypeMove:
				if ub.chatty {
					log.Printf("Touch Moving.\n")
				}
			case touch.TypeEnd:
				if !grabbingVector {
					if ub.chatty {
						log.Printf("That's odd!\n")
					}
				}
				grabbingVector = false
				if ub.chatty {
					log.Printf("Done\n")
					log.Printf("  begin = (%v, %v)\n", ub.beginX, ub.beginY)
					log.Printf("    end = (%v, %v)\n", e.X, e.Y)
					log.Printf("  delta = (%v, %v)\n", e.X-ub.beginX, e.Y-ub.beginY)
				}
				/*
						On X11, screen points come in as some kind of pixels.
						As the screen is resized, 0,0 stays the same,
					  but the other numbers change.

							(0,0)    ... (high, 0)
							...          ...
							(0,high) ... (high, high)

				*/
			}
		case size.Event:
			sz = e
			// These numbers are in the same units as touch events.
			// After a resize,
			//   e.X  <= c.WidthPx
			//   e.Y  <= c.HeightPx
			if ub.chatty {
				log.Printf(" config = (%v, %v)\n", sz.WidthPx, sz.HeightPx)
			}
			ub.touchX = float32(sz.WidthPx / 2)
			ub.touchY = float32(sz.HeightPx / 2)
			// gm.SetOrigin(touchX, touchY)
		}
	}
}
