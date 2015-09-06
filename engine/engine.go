package engine

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
	"math/rand"
)

const (
	// Start with generous slop.
	defaultMaxDistSqForImpulse = 5000
	debugShowResizes           = false
	maxHoldCount               = 30
	magicButtonSideLength      = 100
	fuzzyZero                  = 0.1
	minDragLength              = 6
)

type Engine struct {
	isAlive             bool
	maxDistSqForImpulse float32
	gravity             float32
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
	leftDoor            model.DoorState
	rightDoor           model.DoorState
	chBallCommand       chan model.BallCommand // Owned, written to.
	// A time unit representing how much time (in some unspecified time
	// unit) between each paint event.  Making this number smaller makes
	// balls move faster.
	pauseDuration float32

	// Window height and width are provided in "pixels".
	// Velocity == "pixels traversed per timeStep".
	pixelsToCrossDuringPause float32
}

func NewEngine(
	chatty bool,
	gm *game.V23Manager,
	scn *screen.Screen,
) *Engine {
	if scn == nil {
		log.Panic("Screen cannot be nil")
	}
	if gm == nil {
		log.Panic("V23Manager cannot be nil")
	}
	return &Engine{
		false, // isAlive
		defaultMaxDistSqForImpulse,
		0,     // gravity
		0,     // numBallsCreated
		false, // firstResizeDone
		gm,
		scn,
		chatty,
		[]*model.Ball{},
		0, 0, 0, 0,
		model.Closed, // left door
		model.Closed, // right door
		make(chan model.BallCommand),
		200, // pauseDuration
		20,  // pixelsToCrossDuringPause
	}
}

func (gn *Engine) String() string {
	return fmt.Sprintf("%v %v", gn.gm.Me(), gn.balls)
}

func (gn *Engine) start() {
	if gn.chatty {
		log.Printf("Engine starting.\n")
	}
	gn.scn.Start()

	gn.gm.RunPrep(gn.chBallCommand)
	go gn.gm.Run()

	if gn.firstResizeDone && gn.numBallsCreated < 1 {
		gn.createBall()
	}

	gn.isAlive = true
	if gn.chatty {
		log.Printf("Engine started.\n")
	}
}

func (gn *Engine) stop() {
	gn.scn.Clear()
	if !gn.isAlive {
		log.Println("Stop called on dead gn.")
		return
	}
	if gn.chatty {
		log.Println("****************************** Engine stopping.")
	}
	gn.gm.NoNewBallsOrPeople()
	gn.discardBalls()
	// Closing this channel sends a nil, which has to be handled on the
	// other side - so don't bother to close(gn.chBallCommand)
	if gn.chatty {
		log.Println("Sending shutdown to v23.")
	}
	// Wait for v23 manager to shutdown.
	gn.gm.Stop()
	gn.isAlive = false
	if gn.chatty {
		log.Println("Engine done!")
	}
}

func (gn *Engine) Run(a app.App) {
	if gn.chatty {
		log.Println("Starting gn Run.")
	}
	holdCount := 0
	var sz size.Event
	for {
		select {
		case mc := <-gn.gm.ChMasterCommand():
			switch mc.Name {
			case "kick":
				gn.kick()
			case "freeze":
				gn.freeze()
			case "left":
				gn.left()
			case "right":
				gn.right()
			case "random":
				gn.random()
			case "destroy":
				gn.balls = []*model.Ball{}
			default:
				log.Print("Don't understand command %v", mc)
			}
		case <-gn.gm.ChKick():
			gn.kick()
		case <-gn.gm.ChQuit():
			gn.stop()
			return
		case pd := <-gn.gm.ChPauseDuration():
			gn.pauseDuration = pd
		case g := <-gn.gm.ChGravity():
			gn.gravity = g
		case b := <-gn.gm.ChIncomingBall():
			nx := b.GetPos().X
			if nx == config.MagicX {
				// Ball came in from center of top
				nx = gn.scn.Width() / 2.0
			} else if nx >= 0 && nx <= fuzzyZero {
				// Ball came in from left.
				nx = 0
			} else {
				// Ball came in from right.
				nx = gn.scn.Width()
			}
			// Assume Y component normalized before teleport.
			ny := b.GetPos().Y * gn.scn.Height()
			b.SetPos(nx, ny)
			// TODO: Adjust velocity per refraction-like rules?
			gn.balls = append(gn.balls, b)
		case dc := <-gn.gm.ChDoorCommand():
			gn.handleDoor(dc)
		case event := <-a.Events():
			switch e := app.Filter(event).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					if !gn.isAlive {
						// Calls v23.Init(), determines current players from MT, etc.
						if !gn.gm.IsReadyToRun(false) {
							return
						}
						gn.start()
					}
				case lifecycle.CrossOff:
					gn.stop()
					return
				}
			case paint.Event:
				if !gn.isAlive {
					log.Panic("Not alive, yet told to paint")
				}
				gn.moveBalls()
				gn.scn.Paint(gn.balls)
				a.EndPaint(e)
			case key.Event: // Aspirationally use keys
				if gn.chatty {
					log.Printf("Key event! %T = %v", e.Code, e.Code)
				}
				switch e.Code {
				case key.CodeQ:
					gn.stop()
					return
				case key.CodeEscape:
					gn.stop()
					return
				}
			case touch.Event:
				if gn.chatty {
					log.Println("Touch event")
				}
				switch e.Type {
				case touch.TypeBegin:
					holdCount = 1
					gn.beginX = e.X
					gn.beginY = e.Y
					if e.X < magicButtonSideLength && e.Y < magicButtonSideLength {
						if gn.chatty {
							log.Printf("Touched shutdown spot.\n")
						}
						gn.stop()
						return
					}
				case touch.TypeMove:
					holdCount++
				case touch.TypeEnd:
					if gn.chatty {
						log.Printf("holdcount = %d", holdCount)
					}
					if holdCount > 0 && holdCount <= maxHoldCount {
						// If they hold on too long, ignore it.
						dx := float64(e.X - gn.beginX)
						dy := float64(e.Y - gn.beginY)
						mag := math.Sqrt(dx*dx + dy*dy)
						if mag >= minDragLength {
							ndx := float32(dx/mag) * gn.scn.Width() / gn.pauseDuration
							ndy := float32(dy/mag) * gn.scn.Height() / gn.pauseDuration
							b := model.NewBall(nil,
								model.Vec{gn.beginX, gn.beginY},
								model.Vec{ndx, ndy})
							if gn.chatty {
								log.Printf("Sending impulse: %s", b.String())
							}
							gn.applyImpulse(b)
						} else {
							if gn.chatty {
								log.Printf("Mag only %.4f", mag)
							}
						}
					}
					holdCount = 0
				}
			case size.Event:
				// TODO: Adjust velocity on resizes - balls should take the
				// same amount of time to traverse the screen regardless of
				// the size.
				sz = e
				gn.scn.ReSize(float32(sz.WidthPx), float32(sz.HeightPx))
				gn.resetImpulseLimit()
				if gn.chatty && debugShowResizes {
					log.Printf(
						"Resize new w=%.2f, new h=%.2f, maxDsqImpulse = %f.2",
						gn.scn.Width(),
						gn.scn.Height(),
						gn.maxDistSqForImpulse)
				}
				if !gn.firstResizeDone {
					gn.firstResizeDone = true
					// Don't place the first ball till size is known.
					if gn.numBallsCreated < 1 && gn.gm.IsRunning() {
						gn.createBall()
					}
				}
			}
		}
	}
}

func (gn *Engine) minVelocity() float32 {
	return gn.pixelsToCrossDuringPause / gn.pauseDuration
}

func (gn *Engine) kick() {
	if gn.chatty {
		log.Print("Kicking.")
	}
	for _, b := range gn.balls {
		//	b.SetVel(0, gn.minVelocity())
		b.SetVel(0, gn.scn.Height()/gn.pauseDuration)
	}
}

func (gn *Engine) left() {
	for _, b := range gn.balls {
		b.SetVel(-gn.scn.Width()/gn.pauseDuration, 0)
	}
}

func (gn *Engine) right() {
	for _, b := range gn.balls {
		b.SetVel(gn.scn.Width()/gn.pauseDuration, 0)
	}
}

func (gn *Engine) freeze() {
	if gn.chatty {
		log.Print("Freezing.")
	}
	for _, b := range gn.balls {
		b.SetVel(0, 0)
	}
}

const (
	two  = float64(2.0)
	half = float64(0.5)
)

func randNorm() float64 {
	return two * (rand.Float64() - half)
}

func (gn *Engine) random() {
	if gn.chatty {
		log.Print("Assigning random velocities.")
	}
	coefX := float64(gn.scn.Width() / gn.pauseDuration)
	coefY := float64(gn.scn.Height() / gn.pauseDuration)
	for _, b := range gn.balls {
		b.SetVel(float32(coefX*randNorm()), float32(coefY*randNorm()))
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
func (gn *Engine) moveBalls() {
	discardPile := []discardable{}
	velX0 := gn.scn.Width() / gn.pauseDuration
	velY0 := gn.scn.Height() / gn.pauseDuration
	for i, b := range gn.balls {
		dx := b.GetVel().X
		dy := b.GetVel().Y + gn.gravity

		nx := b.GetPos().X + dx*velX0
		ny := b.GetPos().Y + dy*velY0
		if nx <= 0 {
			// Ball hit left side of screen.
			if gn.leftDoor == model.Open {
				nx = 1
				discardPile = append(discardPile, discardable{i, model.Left})
			} else {
				nx = 0
				dx = -dx
			}
		} else if nx >= gn.scn.Width() {
			// Ball hit right side of screen.
			if gn.rightDoor == model.Open {
				nx = 0
				discardPile = append(discardPile, discardable{i, model.Right})
			} else {
				nx = gn.scn.Width()
				dx = -dx
			}
		}
		if ny <= 0 {
			// Ball hit top of screen.
			ny = 0
			dy = -dy
		} else if ny >= gn.scn.Height() {
			// Ball hit bottom of screen.
			ny = gn.scn.Height()
			dy = -dy
		}
		b.SetPos(nx, ny)
		b.SetVel(dx, dy)
	}
	if gn.chatty {
		if len(discardPile) > 0 {
			log.Printf("%d balls need to move off screen.", len(discardPile))
		}
	}
	gn.throwBalls(discardPile)
}

func (gn *Engine) throwBalls(discardPile []discardable) {
	count := 0
	for _, discard := range discardPile {
		i := discard.i - count
		if gn.chatty {
			log.Printf("Throwing ball %v (i=%d, k=%d, count=%d).\n",
				discard.d, i, discard.i, count)
		}
		count++
		b := gn.balls[i]
		if gn.chatty {
			log.Printf("  ball = %v\n", b)
		}
		gn.balls = append(gn.balls[:i], gn.balls[i+1:]...)
		gn.throwOneBall(b, discard.d)
	}
}

func (gn *Engine) throwOneBall(b *model.Ball, direction model.Direction) {
	// Before throwing, normalize the Y coordinate to a dimensionless
	// percentage.  Recipient converts it based on their own dimensions,
	// so that if the ball left one tenth of the way up the screen, it
	// enters the next screen at the same relative position.
	b.SetPos(b.GetPos().X, b.GetPos().Y/gn.scn.Height())
	gn.chBallCommand <- model.BallCommand{b, direction}
}

type discardable struct {
	i int
	d model.Direction
}

func (gn *Engine) discardBalls() {
	if gn.leftDoor == model.Closed && gn.rightDoor == model.Closed {
		// Nowhere to discard balls.
		return
	}
	minVelocity := gn.minVelocity()
	discardPile := []discardable{}
	for i, b := range gn.balls {
		vx := b.GetVel().X
		vy := b.GetVel().Y
		nx := b.GetPos().X
		if vx < 0 {
			// Try to discard to left.
			if gn.leftDoor == model.Open {
				discardPile = append(discardPile, discardable{i, model.Left})
				// Ball should appear on right side of place it flies to.
				nx = gn.scn.Width()
				if -vx < minVelocity {
					vx = -minVelocity
				}
			} else {
				if gn.rightDoor == model.Open {
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
			if gn.rightDoor == model.Open {
				discardPile = append(discardPile, discardable{i, model.Right})
				nx = 0
				if vx < minVelocity {
					vx = minVelocity
				}
			} else {
				if gn.leftDoor == model.Open {
					discardPile = append(discardPile, discardable{i, model.Left})
					nx = gn.scn.Width()
				}
				// Make ball go left.
				if vx < minVelocity {
					vx = -minVelocity
				} else {
					vx = -vx
				}
			}
		}
		if math.Abs(float64(vy)) < float64(minVelocity) {
			// Kick it up.
			vy = -minVelocity
		}
		b.SetPos(nx, b.GetPos().Y)
		b.SetVel(vx, vy)
	}

	if gn.chatty {
		if len(discardPile) > 0 {
			log.Printf("%d balls to discard.", len(discardPile))
		}
	}
	gn.throwBalls(discardPile)
	gn.balls = []*model.Ball{}
}

func (gn *Engine) createBall() {
	if gn.chatty {
		log.Printf("Creating ball.")
	}
	gn.balls = append(
		gn.balls,
		model.NewBall(
			gn.gm.Me(),
			model.Vec{gn.scn.Width() / 2, gn.scn.Height() / 2},
			model.Vec{0, 0}))
	// Since balls can come in from the outside,  len(gn.balls) is
	// not a reliable indicated of how many balls this code created,
	// so need a distinct counter.
	gn.numBallsCreated++
}

// Use fraction of characteristic screen size
// to define max distance over which an impulse
// is considered to have 'hit' a ball.
func (gn *Engine) resetImpulseLimit() {
	max := gn.scn.Height()
	if gn.scn.Width() > max {
		max = gn.scn.Width()
	}
	max = max / 3
	gn.maxDistSqForImpulse = max * max
}

// Find the ball closest to the impulse and within a reasonable
// range, apply new velocity to the ball.
func (gn *Engine) applyImpulse(impulse *model.Ball) {
	if gn.chatty {
		log.Printf("Got impulse: %s", impulse.String())
	}
	closest, ball := gn.closestDsq(impulse.GetPos())
	if ball == nil {
		if gn.chatty {
			log.Printf("No ball to punch.")
		}
		return
	}
	if gn.chatty {
		log.Printf("DSQ to ball: %f.1\n", closest)
	}
	if closest <= gn.maxDistSqForImpulse {
		if gn.chatty {
			log.Printf("Punching ball.\n")
		}
		ball.SetVel(impulse.GetVel().X, impulse.GetVel().Y)
	} else {
		if gn.chatty {
			log.Printf("Ball further than %f.1\n",
				gn.maxDistSqForImpulse)
		}
	}
}

func (gn *Engine) closestDsq(im model.Vec) (
	smallest float32, target *model.Ball) {
	smallest = math.MaxFloat32
	target = nil
	for _, b := range gn.balls {
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

func (gn *Engine) handleDoor(dc model.DoorCommand) {
	if gn.chatty {
		log.Printf("Received door command: %v", dc)
	}
	if dc.S == model.Open {
		if dc.D == model.Left {
			gn.leftDoor = model.Open
		} else {
			gn.rightDoor = model.Open
		}
	} else {
		if dc.D == model.Left {
			gn.leftDoor = model.Closed
		} else {
			gn.rightDoor = model.Closed
		}
	}
}
