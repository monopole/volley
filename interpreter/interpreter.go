package interpreter

import (
	"fmt"
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/game"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"log"
	"math"
	"net/http"
)

const (
	// Arbitrary - depends on paint event timing.
	timeStep = 10

	// Start with generous slop.
	defaultMaxDistSqForImpulse = 5000

	maxHoldCount          = 20
	magicButtonSideLength = 100
)

type Interpreter struct {
	isAlive             bool
	maxDistSqForImpulse float32
	isGravity           bool
	gm                  *game.V23Manager
	scn                 *screen.Screen
	chatty              bool
	balls               []*model.Ball
	touchX              float32
	touchY              float32
	beginX              float32
	beginY              float32
	leftDoor            model.DoorState
	rightDoor           model.DoorState
	chBallCommand       chan model.BallCommand // Owned, written to.
}

func NewInterpreter(
	chatty bool,
	gm *game.V23Manager,
	scn *screen.Screen,
) *Interpreter {
	if scn == nil {
		log.Panic("Screen cannot be nil")
	}
	if gm == nil {
		log.Panic("V23Manager cannot be nil")
	}
	return &Interpreter{
		false, // isAlive
		defaultMaxDistSqForImpulse,
		false, // isGravity
		gm,
		scn,
		chatty,
		[]*model.Ball{},
		0, 0, 0, 0,
		model.Closed, // left door
		model.Closed, // right door
		make(chan model.BallCommand),
	}
}

func (ub *Interpreter) String() string {
	return fmt.Sprintf("%v %v", ub.gm.Me(), ub.balls)
}

func (ub *Interpreter) quit() {
	if !ub.isAlive {
		log.Println("Quit called on dead interpreter.")
		return
	}
	if ub.chatty {
		log.Println("Interpreter quitting.")
	}
	ub.discardBalls()
	// Closing this channel seems to trigger a ball?
	// close(ub.chBallCommand)
	if ub.chatty {
		log.Println("Sending shutdown to v23.")
	}
	// Wait for v23 manager to shutdown.
	ub.gm.Quit()
	ub.scn.Stop()
	ub.isAlive = false
	if ub.chatty {
		log.Println("Interpreter done.")
	}
}

func gotNetwork() bool {
	_, err := http.Get(config.TestDomain)
	if err == nil {
		log.Printf("Network up - able to hit %s", config.TestDomain)
		return true
	}
	log.Printf("Something wrong with network: %v", err)
	return false
}

func (ub *Interpreter) Run(a app.App) {
	if ub.chatty {
		log.Println("Starting interpreter.")
	}
	holdCount := 0
	var sz size.Event
	for {
		select {
		case b := <-ub.gm.ChIncomingBall():
			if ub.chatty {
				log.Println("Interpreter ball.")
			}
			nx := b.GetPos().X
			if nx <= 0.1 {
				// Fuzzy zero means ball came in from left.
				nx = 0
			} else {
				// Ball came in from right.
				nx = ub.scn.Width()
			}
			// Assume Y component normalized before the throw.
			ny := b.GetPos().Y * ub.scn.Height()
			// Leave the velocity alone for now, although that
			// looks odd when jumping from a small screen to a large screen.
			b.SetPos(nx, ny)
			if ub.chatty {
				log.Printf("Table accepting ball %s", b)
			}
			ub.balls = append(ub.balls, b)
		case dc := <-ub.gm.ChDoorCommand():
			if ub.chatty {
				log.Println("Interpreter received door command.")
			}
			ub.handleDoor(dc)
		case event := <-a.Events():
			switch e := app.Filter(event).(type) {
			case lifecycle.Event:
				if ub.chatty {
					log.Println("Lifecycle event")
				}
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					if !ub.isAlive {
						if config.FailFast {
							if !gotNetwork() {
								return
							}
						}
						// Calls v23.Init(), determines current players from MT, etc.
						if !ub.gm.IsReadyToRun() {
							return
						}
						if ub.chatty {
							log.Printf("App starting!\n")
						}
						ub.makeFirstBall()
						ub.scn.Start()
						go ub.gm.Run(ub.chBallCommand)
						ub.isAlive = true
					}
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
				if !ub.isAlive {
					log.Panic("Not alive, yet told to paint")
				}
				ub.moveBalls()
				ub.scn.Paint(ub.balls)
				a.EndPaint(e)
			case key.Event: // Aspirationally use keys
				if ub.chatty {
					log.Printf("Key event! %T = %v", e.Code, e.Code)
				}
				switch e.Code {
				case key.CodeQ:
					ub.quit()
					return
				case key.CodeEscape:
					ub.quit()
					return
				}
			case touch.Event:
				if ub.chatty {
					log.Println("Touch event")
				}
				switch e.Type {
				case touch.TypeBegin:
					holdCount = 1
					ub.beginX = e.X
					ub.beginY = e.Y
					if e.X < magicButtonSideLength && e.Y < magicButtonSideLength {
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
					if holdCount > 0 && holdCount <= maxHoldCount {
						// If they hold on too long, ignore it.
						dx := float64(e.X - ub.beginX)
						dy := float64(e.Y - ub.beginY)
						mag := math.Sqrt(dx*dx + dy*dy)
						if mag > 10 {
							// They should drag for around 10 pixels.
							b := model.NewBall(nil,
								model.Vec{ub.beginX, ub.beginY},
								model.Vec{float32(dx / mag), float32(dy / mag)})
							if ub.chatty {
								log.Printf("Sending impulse: %s", b.String())
							}
							ub.applyImpulse(b)
						}
					}
					holdCount = 0
				}
			case size.Event:
				if ub.chatty {
					log.Println("ReSize event")
				}
				sz = e
				ub.scn.ReSize(float32(sz.WidthPx), float32(sz.HeightPx))
				ub.resetImpulseLimit()
				if ub.chatty {
					log.Printf(
						"New w=%.2f, new h=%.2f, maxDsqImpulse = %f.2",
						ub.scn.Width(),
						ub.scn.Height(),
						ub.maxDistSqForImpulse)
				}

			}
		}
	}
}

// On X11, screen points come in as some notion of pixels.  As
// the screen is resized, (x,y)==0,0 stays fixed in upper left
// corner.
//   (0,0)      ...  (width, 0)
//   ...             ...
//   (0,height) ...  (width, height)
// A positive y velocity is downward.
// Screen center is (width/2, height/2).
// The width and height come in as integers - but they
// seem to be in the same units (pixels).
func (ub *Interpreter) moveBalls() {
	throwLeft := []int{}
	throwRight := []int{}
	for i, b := range ub.balls {
		dx := b.GetVel().X
		dy := b.GetVel().Y
		if ub.isGravity {
			// TODO
		}
		nx := b.GetPos().X + timeStep*dx
		ny := b.GetPos().Y + timeStep*dy
		if nx <= 0 {
			// Ball hit left side of screen.
			if ub.leftDoor == model.Open {
				nx = 1
				throwLeft = append(throwLeft, i)
			} else {
				nx = 0
				dx = -dx
			}
		} else if nx >= ub.scn.Width() {
			// Ball hit right side of screen.
			if ub.rightDoor == model.Open {
				nx = 0
				throwRight = append(throwRight, i)
			} else {
				nx = ub.scn.Width()
				dx = -dx
			}
		}
		if ny <= 0 {
			// Ball hit top of screen.
			ny = 0
			dy = -dy
		} else if ny >= ub.scn.Height() {
			// Ball hit bottom of screen.
			ny = ub.scn.Height()
			dy = -dy
		}
		b.SetPos(nx, ny)
		b.SetVel(dx, dy)
	}
	ub.throwBalls(throwLeft, throwRight)
}

func (ub *Interpreter) throwBalls(throwLeft, throwRight []int) {
	count := 0
	for _, k := range throwLeft {
		i := k - count
		if ub.chatty {
			log.Printf("Throwing ball left (i=%d, k=%d, count=%d).\n", i, k, count)
		}
		count++
		b := ub.balls[i]
		ub.balls = append(ub.balls[:i], ub.balls[i+1:]...)
		ub.throwOneBall(b, model.Left)
	}
	for _, k := range throwRight {
		i := k - count
		if ub.chatty {
			log.Printf("Throwing ball right (i=%d, k=%d, count=%d).\n", i, k, count)
		}
		count++
		b := ub.balls[i]
		ub.balls = append(ub.balls[:i], ub.balls[i+1:]...)
		ub.throwOneBall(b, model.Right)
	}
}

func (ub *Interpreter) throwOneBall(b *model.Ball, direction model.Direction) {
	// Before throwing, normalize the Y coordinate to a dimensionless
	// percentage.  Recipient converts it based on their own dimensions,
	// so that if the ball left one tenth of the way up the screen, it
	// enters the next screen at the same relative position.
	b.SetPos(b.GetPos().X, b.GetPos().Y/ub.scn.Height())
	ub.chBallCommand <- model.BallCommand{b, direction}
}

func (ub *Interpreter) discardBalls() {
	if ub.leftDoor == model.Closed && ub.rightDoor == model.Closed {
		// Nowhere to discard balls.
		return
	}
	throwLeft := []int{}
	throwRight := []int{}
	for i, b := range ub.balls {
		nx := b.GetPos().X
		if b.GetVel().X <= 0 {
			if ub.leftDoor == model.Open {
				throwLeft = append(throwLeft, i)
				// Ball should appear on right side of ub it flies to.
				nx = ub.scn.Width()
			} else {
				if ub.rightDoor == model.Open {
					throwRight = append(throwRight, i)
					nx = 0
				}
			}
		} else {
			if ub.rightDoor == model.Open {
				throwRight = append(throwRight, i)
				nx = 0
			} else {
				if ub.leftDoor == model.Open {
					throwLeft = append(throwLeft, i)
					nx = ub.scn.Width()
				}
			}
		}
		b.SetPos(nx, b.GetPos().Y)
	}
	ub.throwBalls(throwLeft, throwRight)
}

func (ub *Interpreter) makeFirstBall() {
	if ub.chatty {
		log.Printf("Making the first ball.")
	}
	if len(ub.balls) > 0 {
		log.Panic("There should be zero balls now.")
	}
	ub.balls = append(
		ub.balls,
		model.NewBall(
			ub.gm.Me(),
			model.Vec{ub.scn.Width() / 2, ub.scn.Height() / 2},
			model.Vec{0, 0}))
}

// Use fraction of characteristic screen size
// to define max distance over which an impulse
// is considered to have 'hit' a ball.
func (ub *Interpreter) resetImpulseLimit() {
	max := ub.scn.Height()
	if ub.scn.Width() > max {
		max = ub.scn.Width()
	}
	max = max / 4
	ub.maxDistSqForImpulse = max * max
}

func (ub *Interpreter) applyImpulse(impulse *model.Ball) {
	// Find the ball closest to the impulse and
	// within a reasonable range,
	// apply new velocity to the ball.
	// For now, just pick the zero ball.
	if ub.chatty {
		log.Printf("Got impulse: %s", impulse.String())
	}
	closest, ball := ub.closestDsq(impulse.GetPos())
	if ball == nil {
		if ub.chatty {
			log.Printf("No ball to punch.")
		}
		return
	}
	if ub.chatty {
		log.Printf("DSQ to ball: %f.1\n", closest)
	}
	if closest <= ub.maxDistSqForImpulse {
		if ub.chatty {
			log.Printf("Punching ball.\n")
		}
		ball.SetVel(impulse.GetVel().X, impulse.GetVel().Y)
	} else {
		if ub.chatty {
			log.Printf("Ball further than %f.1\n",
				ub.maxDistSqForImpulse)
		}
	}
}

func (ub *Interpreter) closestDsq(im model.Vec) (
	smallest float32, target *model.Ball) {
	smallest = math.MaxFloat32
	target = nil
	for _, b := range ub.balls {
		dx := im.X - b.GetPos().X
		dy := im.Y - b.GetPos().Y
		dsq := dx*dx + dy*dy
		if dsq < smallest {
			smallest = dsq
			target = b
		}
	}
	return
}

func (ub *Interpreter) handleDoor(dc model.DoorCommand) {
	if ub.chatty {
		log.Printf("Received door command: %v", dc)
	}
	if dc.S == model.Open {
		if dc.D == model.Left {
			ub.leftDoor = model.Open
		} else {
			ub.rightDoor = model.Open
		}
	} else {
		if dc.D == model.Left {
			ub.leftDoor = model.Closed
		} else {
			ub.rightDoor = model.Closed
		}
	}
}
