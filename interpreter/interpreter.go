package interpreter

import (
	"github.com/monopole/croupier/model"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"log"
	"math"
)

type Interpreter struct {
	chatty        bool
	chQuit        chan<- chan bool         // Not owned, written to.
	chExecCommand chan<- model.ExecCommand // Not owned, written to.
	chImpulse     chan<- *model.Ball       // Not owned, written to.
	chResize      chan<- model.Vec         // Not owned, written to.
	touchX        float32
	touchY        float32
	beginX        float32
	beginY        float32
}

func NewInterpreter(
	chatty bool,
	chQuit chan<- chan bool,
	chExecCommand chan<- model.ExecCommand,
	chImpulse chan<- *model.Ball,
	chResize chan<- model.Vec,
) *Interpreter {
	return &Interpreter{
		chatty,
		chQuit,
		chExecCommand,
		chImpulse,
		chResize,
		0, 0, 0, 0,
	}
}

func (ub *Interpreter) quit() {
	ub.chExecCommand <- model.ExecStop
	ch := make(chan bool)
	ub.chQuit <- ch
	<-ch
}

func (ub *Interpreter) Run(a app.App) {
	if ub.chatty {
		log.Println("Starting interpreter.")
	}
	holdCount := 0
	var sz size.Event
	for e := range a.Events() {
		switch e := app.Filter(e).(type) {
		case lifecycle.Event:
			switch e.Crosses(lifecycle.StageVisible) {
			case lifecycle.CrossOn:
				if ub.chatty {
					log.Printf("App starting!\n")
				}
				ub.chExecCommand <- model.ExecStart
			case lifecycle.CrossOff:
				if ub.chatty {
					log.Printf("App stopping!\n")
				}
				// TODO(monopole): Perhaps there's a mode where the screen
				// is stopped but the app keeps going?
				ub.quit()
				return
			}
		case paint.Event:
			ub.chExecCommand <- model.ExecPaint
			// TODO(jregan): might have to wait for a response
			// before calling back to the app.
			a.EndPaint(e)
		case touch.Event:
			switch e.Type {
			case touch.TypeBegin:
				holdCount = 1
				if ub.chatty {
					// 	log.Printf("Touch Begin.\n")
				}
				ub.beginX = e.X
				ub.beginY = e.Y
				if e.X < 10 && e.Y < 10 {
					if ub.chatty {
						log.Printf("Touched shutdown spot.\n")
					}
					ub.quit()
					return
				}
			case touch.TypeMove:
				holdCount++
				if ub.chatty {
					// log.Printf("Touch Moving.\n")
				}
			case touch.TypeEnd:
				if holdCount > 0 && holdCount < 20 {
					// If they hold on too long, ignore it.
					dx := float64(e.X - ub.beginX)
					dy := float64(e.Y - ub.beginY)
					mag := math.Sqrt(dx*dx + dy*dy)
					if mag > 10 {
						// They should drag for around 10 pixels.
						b := model.NewBall(nil,
							model.Vec{ub.beginX, ub.beginY},
							model.Vec{float32(dx / mag), float32(dy / mag)})
						ub.chImpulse <- b
					}
				}
				holdCount = 0

				/*
										On X11, screen points come in as some kind of pixels.
										As the screen is resized, 0,0 stays the same,
										but the other numbers change.

										(0,0)    ... (high, 0)
										...          ...
										(0,high) ... (high, high)

										After a resize, the center of the screen is
										x = float32(sz.WidthPx / 2)
										y = float32(sz.HeightPx / 2)

					          The width and height come in as integers - but they
					          seem to be in the same units (pixels).

				*/
			}
		case size.Event:
			sz = e
			ub.chResize <- model.Vec{float32(sz.WidthPx), float32(sz.HeightPx)}
		}
	}
}
