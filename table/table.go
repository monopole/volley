package table

import (
	"fmt"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"log"
)

type Table struct {
	chatty        bool
	me            *model.Player
	scn           *screen.Screen
	screenOn      bool
	chBallEnter   <-chan *model.Ball       // Not owned, read from.
	chDoorCommand <-chan model.DoorCommand // Not owned, read from.
	chV23         chan<- chan bool         // Not owned, written to.
	chExecCommand chan model.ExecCommand   // Owned, read from.
	chImpulse     chan *model.Ball         // Owned, read from.
	chResize      chan model.Vec           // Owned, read from.
	chQuit        chan chan bool           // Owned, read from.
	chBallCommand chan model.BallCommand   // Owned, written to.
	balls         []*model.Ball
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
				log.Printf("Got resize: %v", rs)
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
			if len(table.balls) > 0 {
				b := table.balls[0]
				b.SetVel(impulse.GetVel().X, impulse.GetVel().Y)
			}
		}
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
