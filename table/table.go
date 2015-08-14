package table

import (
	"fmt"
	"github.com/monopole/croupier/model"
	"log"
)

type commandType int

const (
	commandError commandType = iota
	commandQuit
	commandRandomImpulse
	commandImpulse
)

type Table struct {
	me              *model.Player
	commands        <-chan commandType
	chBallEnter     <-chan *model.Ball
	chBallExitLeft  chan<- *model.Ball
	chBallExitRight chan<- *model.Ball
	balls           []*model.Ball
}

func NewTable(
	me *model.Player,
	commands <-chan commandType,
	chBallEnter <-chan *model.Ball,
	chBallExitLeft chan<- *model.Ball,
	chBallExitRight chan<- *model.Ball,
) *Table {
	return &Table{me,
		commands, chBallEnter, chBallExitLeft, chBallExitRight,
		[]*model.Ball{model.NewBall(me, model.Vec{0, 0}, model.Vec{0, 0})}}
}

func (table *Table) String() string {
	return fmt.Sprintf("%v %v", table.me, table.balls)
}

func (table *Table) play() {
	log.Println("play entered.")
	for {
		log.Println("play loop zz top.")
		select {
		case c := <-table.commands:
			switch c {
			case commandRandomImpulse:
			case commandImpulse:
			case commandQuit:
			}
		case b := <-table.chBallEnter:
			table.balls = append(table.balls, b)
		}
	}
}

func (table *Table) Quit() {
}
