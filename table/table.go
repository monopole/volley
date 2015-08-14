package table

import (
	"fmt"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/screen"
	"log"
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
	chatty          bool
	screen          *screen.Screen
	commands        <-chan commandType
	chImpulse       <-chan *model.Ball
	chBallEnter     <-chan *model.Ball
	chBallExitLeft  chan<- *model.Ball
	chBallExitRight chan<- *model.Ball
	chQuit          chan chan bool
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
	return &Table{me, true, s,
		commands,
		chImpulse, chBallEnter, chBallExitLeft, chBallExitRight,
		make(chan chan bool),
		[]*model.Ball{model.NewBall(me, model.Vec{0, 0}, model.Vec{0, 0})}}
}

func (table *Table) Quitter() chan<- chan bool {
	return table.chQuit
}

func (table *Table) String() string {
	return fmt.Sprintf("%v %v", table.me, table.balls)
}

func (table *Table) play() {
	for {
		select {
		case ch := <-table.chQuit:
			table.quit()
			ch <- true
			return
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
			// Find the ball closest to the impulse and within a reasonable range,
			// apply new velocity to the ball.
			// For now, just pick the zero ball.
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
}
