package table

import (
	"fmt"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"log"
	"math"
)

// Arbitrary - depends on paint event timing.
const timeStep = 2

type Table struct {
	chatty              bool
	maxDistSqForImpulse float32
	me                  *model.Player
	scn                 *screen.Screen
	screenOn            bool
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
		5000, // maxDistSqForImpulse, start generously
		me, scn,
		false, // screenOn
		chBallEnter,
		chDoorCommand,
		chV23,
		make(chan model.ExecCommand),
		make(chan *model.Ball),
		make(chan model.Vec),
		make(chan chan bool),
		make(chan model.BallCommand),
		// Every player starts with their own ball.
		[]*model.Ball{model.NewBall(me, model.Vec{0, 0}, model.Vec{0, 0})}}
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
			if table.chatty {
				log.Printf("Received door command: %v", dc)
			}
		case rs := <-table.chResize:
			if table.chatty {
			}
			min := rs.Y
			if rs.X < rs.Y {
				min = rs.X
			}
			// Take a fraction of that characteristic distance,
			// and use it to define impulse proximity.
			min = min / 8
			table.maxDistSqForImpulse = min * min
			if table.chatty {
				log.Printf(
					"Got resize: %v, maxDsqImpulse = %f.2",
					rs, table.maxDistSqForImpulse)
			}
			table.scn.ReSize(rs.X, rs.Y)
			if len(table.balls) > 0 {
				table.balls[0].SetPos(rs.X/2, rs.Y/2)
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
			case model.ExecStop:
				if table.scn != nil && table.screenOn {
					table.scn.Stop()
				}
			}
		case impulse := <-table.chImpulse:
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
				break
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
	}
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

// Screen coordinate system is (x,y)
//
//   (0,0)    ... (high, 0)
//   ...          ...
//   (0,high) ... (high, high)
//
// A positive y velocity is downward.
//
func (table *Table) moveBalls() {
	for _, b := range table.balls {
		dx := b.GetVel().X
		dy := b.GetVel().Y
		nx := b.GetPos().X + timeStep*dx
		ny := b.GetPos().Y + timeStep*dy
		if nx <= 0 {
			// Ball hit left side of screen.
			nx = 0
			dx = -dx
		} else if nx >= table.scn.Width() {
			// Ball hit right side of screen.
			nx = table.scn.Width()
			dx = -dx
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
}

func (table *Table) quit() {
	if table.chatty {
		log.Println("Table quitting.")
	}
	// Closing this channel seems to trigger a ball?
	// close(table.chBallCommand)
	if table.chatty {
		log.Println("Sending shutdown to v23.")
	}
	ch := make(chan bool)
	table.chV23 <- ch
	<-ch
}
