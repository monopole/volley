package interpreter

import (
	"fmt"
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
)

const (
	// Start with generous slop.
	defaultMaxDistSqForImpulse = 5000
	debugShowResizes           = false
	maxHoldCount               = 20
	magicButtonSideLength      = 100
	fuzzyZero                  = 0.1
	minDragLength              = 10
	// Arbitrary statement that this many time units passed between each
	// paint event.  Making this number smaller makes balls move faster.
	timeStep = 100.0
	// Window height and width are provided in "pixels".
	// Velocity == "pixels traversed per timeStep".
	minVelocity = 20.0 / timeStep
)

type Interpreter struct {
	isAlive             bool
	maxDistSqForImpulse float32
	isGravity           bool
	numBallsCreated     int
	firstResizeDone     bool
	gm                  *game.V23Manager
	scn                 *screen.Screen
	chatty              bool
	balls               []*model.Ball
	touchX              float32
	touchY              float32
	beginX              float32
	beginY              float32
	velocityX           float32
	velocityY           float32
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
		0,     // numBallsCreated
		false, // firstResizeDone
		gm,
		scn,
		chatty,
		[]*model.Ball{},
		0, 0, 0, 0, minVelocity, minVelocity,
		model.Closed, // left door
		model.Closed, // right door
		make(chan model.BallCommand),
	}
}

func (ub *Interpreter) String() string {
	return fmt.Sprintf("%v %v", ub.gm.Me(), ub.balls)
}

func (ub *Interpreter) start() {
	if ub.chatty {
		log.Printf("Interpreter starting.\n")
	}
	ub.scn.Start()

	ub.gm.RunPrep(ub.chBallCommand)
	go ub.gm.Run()

	if ub.firstResizeDone && ub.numBallsCreated < 1 {
		ub.createBall()
	}

	ub.isAlive = true
	if ub.chatty {
		log.Printf("Interpreter started.\n")
	}
}

func (ub *Interpreter) stop() {
	if !ub.isAlive {
		log.Println("Stop called on dead interpreter.")
		return
	}
	if ub.chatty {
		log.Println("****************************** Interpreter stopping.")
	}
	ub.gm.NoNewBallsOrPeople()
	ub.discardBalls()
	// Closing this channel sends a nil, which has to be handled on the
	// other side - so don't bother to close(ub.chBallCommand)
	if ub.chatty {
		log.Println("Sending shutdown to v23.")
	}
	// Wait for v23 manager to shutdown.
	ub.gm.Stop()
	ub.scn.Stop()
	ub.isAlive = false
	if ub.chatty {
		log.Println("Interpreter done!")
	}
}

func (ub *Interpreter) Run(a app.App) {
	if ub.chatty {
		log.Println("Starting interpreter Run.")
	}
	holdCount := 0
	var sz size.Event
	for {
		select {
		case b := <-ub.gm.ChIncomingBall():
			nx := b.GetPos().X
			if nx <= fuzzyZero {
				// Ball came in from left.
				nx = 0
			} else {
				// Ball came in from right.
				nx = ub.scn.Width()
			}
			// Assume Y component normalized before teleport.
			ny := b.GetPos().Y * ub.scn.Height()
			b.SetPos(nx, ny)
			// TODO: Adjust velocity per refraction-like rules?
			ub.balls = append(ub.balls, b)
		case dc := <-ub.gm.ChDoorCommand():
			ub.handleDoor(dc)
		case event := <-a.Events():
			switch e := app.Filter(event).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					if !ub.isAlive {
						// Calls v23.Init(), determines current players from MT, etc.
						if !ub.gm.IsReadyToRun() {
							return
						}
						ub.start()
					}
				case lifecycle.CrossOff:
					ub.stop()
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
					ub.stop()
					return
				case key.CodeEscape:
					ub.stop()
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
						ub.stop()
						return
					}
				case touch.TypeMove:
					holdCount++
				case touch.TypeEnd:
					if holdCount > 0 && holdCount <= maxHoldCount {
						// If they hold on too long, ignore it.
						dx := float64(e.X - ub.beginX)
						dy := float64(e.Y - ub.beginY)
						mag := math.Sqrt(dx*dx + dy*dy)
						if mag >= minDragLength {
							b := model.NewBall(nil,
								model.Vec{ub.beginX, ub.beginY},
								// Ball velocities differ only in direction
								// at the moment.
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
				// TODO: Adjust velocity on resizes - balls should take the
				// same amount of time to traverse the screen regardless of
				// the size.
				sz = e
				ub.scn.ReSize(float32(sz.WidthPx), float32(sz.HeightPx))
				ub.velocityX = ub.scn.Width() / timeStep
				ub.velocityY = ub.scn.Height() / timeStep
				ub.resetImpulseLimit()
				if ub.chatty && debugShowResizes {
					log.Printf(
						"Resize new w=%.2f, new h=%.2f, maxDsqImpulse = %f.2",
						ub.scn.Width(),
						ub.scn.Height(),
						ub.maxDistSqForImpulse)
				}
				if !ub.firstResizeDone {
					ub.firstResizeDone = true
					// Don't place the first ball till size is known.
					if ub.numBallsCreated < 1 && ub.gm.IsRunning() {
						ub.createBall()
					}
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
	discardPile := []discardable{}
	for i, b := range ub.balls {
		dx := b.GetVel().X
		dy := b.GetVel().Y
		if ub.isGravity {
			// TODO
		}
		nx := b.GetPos().X + ub.velocityX*dx
		ny := b.GetPos().Y + ub.velocityY*dy
		if nx <= 0 {
			// Ball hit left side of screen.
			if ub.leftDoor == model.Open {
				nx = 1
				discardPile = append(discardPile, discardable{i, model.Left})
			} else {
				nx = 0
				dx = -dx
			}
		} else if nx >= ub.scn.Width() {
			// Ball hit right side of screen.
			if ub.rightDoor == model.Open {
				nx = 0
				discardPile = append(discardPile, discardable{i, model.Right})
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
	if ub.chatty {
		if len(discardPile) > 0 {
			log.Printf("%d balls need to move off screen.", len(discardPile))
		}
	}
	ub.throwBalls(discardPile)
}

func (ub *Interpreter) throwBalls(discardPile []discardable) {
	count := 0
	for _, discard := range discardPile {
		i := discard.i - count
		if ub.chatty {
			log.Printf("Throwing ball %v (i=%d, k=%d, count=%d).\n",
				discard.d, i, discard.i, count)
		}
		count++
		b := ub.balls[i]
		if ub.chatty {
			log.Printf("  ball = %v\n", b)
		}
		ub.balls = append(ub.balls[:i], ub.balls[i+1:]...)
		ub.throwOneBall(b, discard.d)
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

type discardable struct {
	i int
	d model.Direction
}

func (ub *Interpreter) discardBalls() {
	if ub.leftDoor == model.Closed && ub.rightDoor == model.Closed {
		// Nowhere to discard balls.
		return
	}
	discardPile := []discardable{}
	for i, b := range ub.balls {
		vx := b.GetVel().X
		vy := b.GetVel().Y
		nx := b.GetPos().X
		if vx < 0 {
			// Try to discard to left.
			if ub.leftDoor == model.Open {
				discardPile = append(discardPile, discardable{i, model.Left})
				// Ball should appear on right side of place it flies to.
				nx = ub.scn.Width()
				if -vx < minVelocity {
					vx = -minVelocity
				}
			} else {
				if ub.rightDoor == model.Open {
					discardPile = append(discardPile, discardable{i, model.Right})
					nx = 0
					// Make ball go right.
					if -vx < minVelocity {
						vx = minVelocity
					} else {
						vx = -vx
					}
				}
			}
		} else {
			// vx non-negative, try to discard right.
			if ub.rightDoor == model.Open {
				discardPile = append(discardPile, discardable{i, model.Right})
				nx = 0
				if vx < minVelocity {
					vx = minVelocity
				}
			} else {
				if ub.leftDoor == model.Open {
					discardPile = append(discardPile, discardable{i, model.Left})
					nx = ub.scn.Width()
				}
				// Make ball go left.
				if vx < minVelocity {
					vx = -minVelocity
				} else {
					vx = -vx
				}
			}
		}
		if math.Abs(float64(vy)) < minVelocity {
			vy = minVelocity
		}
		b.SetPos(nx, b.GetPos().Y)
		b.SetVel(vx, vy)
	}

	if ub.chatty {
		if len(discardPile) > 0 {
			log.Printf("%d balls to discard.", len(discardPile))
		}
	}
	ub.throwBalls(discardPile)
}

func (ub *Interpreter) createBall() {
	if ub.chatty {
		log.Printf("Creating ball.")
	}
	ub.balls = append(
		ub.balls,
		model.NewBall(
			ub.gm.Me(),
			model.Vec{ub.scn.Width() / 2, ub.scn.Height() / 2},
			model.Vec{0, 0}))
	// Since balls can come in from the outside,  len(ub.balls) is
	// not a reliable indicated of how many balls this thread created,
	// so need a distinct counter.
	ub.numBallsCreated++
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

// Find the ball closest to the impulse and within a reasonable
// range, apply new velocity to the ball.
func (ub *Interpreter) applyImpulse(impulse *model.Ball) {
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
