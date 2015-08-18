package table

import (
	"fmt"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"log"
	"math"
)

// Arbitrary - depends on paint event timing.
const timeStep = 10

// Start with generous slop.
const defaultMaxDistSqForImpulse = 5000

type Table struct {
	chatty              bool
	maxDistSqForImpulse float32
	me                  *model.Player
	scn                 *screen.Screen
	screenOn            bool
	isGravity           bool
	isLeftDoorOpen      bool
	isRightDoorOpen     bool
	chBallEnter         <-chan *model.Ball       // Not owned, read from.
	chDoorCommand       <-chan model.DoorCommand // Not owned, read from.
	chV23               chan<- chan bool         // Not owned, written to.
	chExecCommand       chan model.ExecCommand   // Owned, read from.
	chImpulse           chan *model.Ball         // Owned, read from.
	chResize            chan model.Vec           // Owned, read from.
	chQuit              chan chan bool           // Owned, read from.
	chBallCommand       chan model.BallCommand   // Owned, written to.
	balls               []*model.Ball
}

func NewTable(
	chatty bool,
	me *model.Player,
	scn *screen.Screen,
	chBallEnter <-chan *model.Ball,
	chDoorCommand <-chan model.DoorCommand,
	chV23 chan<- chan bool,
) *Table {
	return &Table{
		chatty,
		defaultMaxDistSqForImpulse,
		me, scn,
		false, // screenOn
		false, // isGravity
		false, //	isLeftDoorOpen
		false, //	isRightDoorOpen
		chBallEnter,
		chDoorCommand,
		chV23,
		make(chan model.ExecCommand),
		make(chan *model.Ball),
		make(chan model.Vec),
		make(chan chan bool),
		make(chan model.BallCommand),
		[]*model.Ball{},
	}
}

func (table *Table) ChBallCommand() <-chan model.BallCommand {
	return table.chBallCommand
}

func (table *Table) ChExecCommand() chan<- model.ExecCommand {
	return table.chExecCommand
}

func (table *Table) ChImpulse() chan<- *model.Ball {
	return table.chImpulse
}

func (table *Table) ChResize() chan<- model.Vec {
	return table.chResize
}

func (table *Table) ChQuit() chan<- chan bool {
	return table.chQuit
}

func (table *Table) String() string {
	return fmt.Sprintf("%v %v", table.me, table.balls)
}

func (table *Table) Run() {
	if table.chatty {
		log.Println("Starting table run.")
	}
	for {
		select {
		case ch := <-table.chQuit:
			table.quit()
			ch <- true
			return
		case b := <-table.chBallEnter:
			table.balls = append(table.balls, b)
		case dc := <-table.chDoorCommand:
			table.handleDoor(dc)
		case rs := <-table.chResize:
			if table.scn.Width() < 1 && table.me.Id() == 1 {
				// Special case: my screen has not resized/rendered yet, and
				// i'm the first player.
				table.firstBall(rs)
			}
			table.scn.ReSize(rs.X, rs.Y)
			table.resetImpulseLimit()
			if table.chatty {
				log.Printf(
					"resize: %v, maxDsqImpulse = %f.2",
					rs, table.maxDistSqForImpulse)
			}
		case c := <-table.chExecCommand:
			switch c {
			case model.ExecRandomImpulse:
			case model.ExecStart:
				if table.scn != nil && !table.screenOn {
					table.scn.Start()
					table.screenOn = true
				}
			case model.ExecPaint:
				if table.scn != nil && table.screenOn {
					table.moveBalls()
					table.scn.Paint(table.balls)
				}
			}
		case impulse := <-table.chImpulse:
			table.applyImpulse(impulse)
		}
	}
}

func (table *Table) firstBall(rs model.Vec) {
	if table.chatty {
		log.Printf("Making the first ball.")
	}
	if len(table.balls) > 0 {
		log.Panic("There should be zero balls now.")
	}
	table.balls = append(
		table.balls,
		model.NewBall(
			table.me,
			model.Vec{rs.X / 2, rs.Y / 2},
			model.Vec{0, 0}))
}

func (table *Table) handleDoor(dc model.DoorCommand) {
	if table.chatty {
		log.Printf("Received door command: %v", dc)
	}
	if dc.S == model.Open {
		if dc.D == model.Left {
			table.isLeftDoorOpen = true
		} else {
			table.isRightDoorOpen = true
		}
	} else {
		if dc.D == model.Left {
			table.isLeftDoorOpen = false
		} else {
			table.isRightDoorOpen = false
		}
	}
}

func (table *Table) applyImpulse(impulse *model.Ball) {
	// Find the ball closest to the impulse and
	// within a reasonable range,
	// apply new velocity to the ball.
	// For now, just pick the zero ball.
	if table.chatty {
		log.Printf("Got impulse: %v", impulse)
	}
	closest, ball := table.closestDsq(impulse.GetPos())
	if ball == nil {
		if table.chatty {
			log.Printf("No ball to punch.")
		}
		return
	}
	if table.chatty {
		log.Printf("DSQ to ball: %f.1\n", closest)
	}
	if closest <= table.maxDistSqForImpulse {
		if table.chatty {
			log.Printf("Punching ball.\n")
		}
		ball.SetVel(impulse.GetVel().X, impulse.GetVel().Y)
	} else {
		if table.chatty {
			log.Printf("Ball further than %f.1\n",
				table.maxDistSqForImpulse)
		}
	}
}

// Use fraction of characteristic screen size
// to define max distance over which an impulse
// is considered to have 'hit' a ball.
func (table *Table) resetImpulseLimit() {
	max := table.scn.Height()
	if table.scn.Width() > max {
		max = table.scn.Width()
	}
	max = max / 4
	table.maxDistSqForImpulse = max * max
}

func (table *Table) closestDsq(im model.Vec) (
	smallest float32, target *model.Ball) {
	smallest = math.MaxFloat32
	target = nil
	for _, b := range table.balls {
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
func (table *Table) moveBalls() {
	throwLeft := []int{}
	throwRight := []int{}
	for i, b := range table.balls {
		dx := b.GetVel().X
		dy := b.GetVel().Y
		if table.isGravity {
			// TODO
		}
		nx := b.GetPos().X + timeStep*dx
		ny := b.GetPos().Y + timeStep*dy
		if nx <= 0 {
			// Ball hit left side of screen.
			if table.isLeftDoorOpen {
				throwLeft = append(throwLeft, i)
				nx = table.scn.Width()
			} else {
				nx = 0
				dx = -dx
			}
		} else if nx >= table.scn.Width() {
			// Ball hit right side of screen.
			if table.isRightDoorOpen {
				throwRight = append(throwRight, i)
				nx = 0
			} else {
				nx = table.scn.Width()
				dx = -dx
			}
		}
		if ny <= 0 {
			// Ball hit top of screen.
			ny = 0
			dy = -dy
		} else if ny >= table.scn.Height() {
			// Ball hit bottom of screen.
			ny = table.scn.Height()
			dy = -dy
		}
		b.SetPos(nx, ny)
		b.SetVel(dx, dy)
	}
	table.throwBalls(throwLeft, throwRight)
}

func (table *Table) discardBalls() {
	if !table.isLeftDoorOpen && !table.isRightDoorOpen {
		// Nowhere to discard balls.
		return
	}
	throwLeft := []int{}
	throwRight := []int{}
	for i, b := range table.balls {
		nx := b.GetPos().X
		if b.GetVel().X <= 0 {
			if table.isLeftDoorOpen {
				throwLeft = append(throwLeft, i)
				nx = table.scn.Width()
			} else {
				if table.isRightDoorOpen {
					throwRight = append(throwRight, i)
					nx = 0
				}
			}
		} else {
			if table.isRightDoorOpen {
				throwRight = append(throwRight, i)
				nx = 0
			} else {
				if table.isLeftDoorOpen {
					throwLeft = append(throwLeft, i)
					nx = table.scn.Width()
				}
			}
		}
		b.SetPos(nx, b.GetPos().Y)
	}
	table.throwBalls(throwLeft, throwRight)
}

func (table *Table) throwBalls(throwLeft, throwRight []int) {
	count := 0
	for _, k := range throwLeft {
		i := k - count
		if table.chatty {
			log.Printf("Throwing ball left (i=%d, k=%d, count=%d).\n", i, k, count)
		}
		count++
		b := table.balls[i]
		table.balls = append(table.balls[:i], table.balls[i+1:]...)
		table.chBallCommand <- model.BallCommand{b, model.Left}
	}
	for _, k := range throwRight {
		i := k - count
		if table.chatty {
			log.Printf("Throwing ball right (i=%d, k=%d, count=%d).\n", i, k, count)
		}
		count++
		b := table.balls[i]
		table.balls = append(table.balls[:i], table.balls[i+1:]...)
		table.chBallCommand <- model.BallCommand{b, model.Right}
	}
}

func (table *Table) quit() {
	if table.chatty {
		log.Println("Table quitting.")
	}
	table.discardBalls()
	// Closing this channel seems to trigger a ball?
	// close(table.chBallCommand)
	if table.chatty {
		log.Println("Sending shutdown to v23.")
	}
	// Wait for the v23 manager to shutdown.
	ch := make(chan bool)
	table.chV23 <- ch
	<-ch
	if table.scn != nil && table.screenOn {
		table.scn.Stop()
	}
}
