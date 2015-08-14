package table

import (
	"fmt"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
)

type commandType int

const (
	commandError commandType = iota
	commandStart
	commandStop
	commandRandomImpulse
	commandPaint
)

type Table struct {
	me              *model.Player
	screen          *screen.Screen
	commands        <-chan commandType
	chImpulse       <-chan *model.Ball
	chBallEnter     <-chan *model.Ball
	chBallExitLeft  chan<- *model.Ball
	chBallExitRight chan<- *model.Ball
	balls           []*model.Ball
}

func NewTable(
	me *model.Player,
	s *screen.Screen,
	commands <-chan commandType,
	chImpulse <-chan *model.Ball,
	chBallEnter <-chan *model.Ball,
	chBallExitLeft chan<- *model.Ball,
	chBallExitRight chan<- *model.Ball,
) *Table {
	return &Table{me, s,
		commands, chImpulse, chBallEnter, chBallExitLeft, chBallExitRight,
		[]*model.Ball{model.NewBall(me, model.Vec{0, 0}, model.Vec{0, 0})}}
}

func (table *Table) String() string {
	return fmt.Sprintf("%v %v", table.me, table.balls)
}

func (table *Table) play() {
	for {
		select {
		case c := <-table.commands:
			switch c {
			case commandRandomImpulse:
			case commandPaint:
				table.screen.Paint(table.balls)
			case commandStart:
				table.screen.Start()
			case commandStop:
				table.screen.Stop()
			}
		case b := <-table.chBallEnter:
			table.balls = append(table.balls, b)
		case impulse := <-table.chImpulse:
			// Find the ball closest to the impulse,
			// apply the new velocity to the ball.
			// For now, just pick the zero ball.
			if len(table.balls) > 0 {
				b := table.balls[0]
				b.SetVel(impulse.GetVel().X, impulse.GetVel().Y)
			}
		}
	}
}

func (table *Table) Quit() {
}
