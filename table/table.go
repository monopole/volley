package table

import (
	"fmt"
	"github.com/monopole/croupier/model"
)

type commandType int

const (
	commandError commandType = iota
	commandQuit
	commandRandomImpulse
	commandImpulse
)

type Table struct {
	me           *model.Player
	balls        []*model.Ball
	players      []*model.Player
	quitting     chan bool
	isDone       chan bool
	commands     chan commandType
	addPlayer    chan *model.Player
	removePlayer chan int
}

func NewTable(
	id int,
	isDone chan bool,
	commands chan commandType,
	addPlayer chan *model.Player,
	removePlayer chan int) *Table {

	me := model.NewPlayer(id)

	return &Table{me,
		[]*model.Ball{model.NewBall(me)},
		[]*model.Player{me},
		make(chan bool),
		isDone,
		commands,
		addPlayer,
		removePlayer}
}

func (table *Table) String() string {
	return fmt.Sprintf("%s %v", table.me, table.players)
}

func (table *Table) play() {
	fmt.Println("play entered.")
	for {
		fmt.Println("play loop zz top.")
		select {
		case <-table.quitting:
			table.isDone <- true
			fmt.Println("All done.")
			return
		case c := <-table.commands:
			switch c {
			case commandRandomImpulse:
			case commandImpulse:
			case commandQuit:

			}
		case player := <-table.addPlayer:
			table.players = append(table.players, player)
		case playerNumber := <-table.removePlayer:
			fmt.Println("TODO removePlayer %v", playerNumber)
		}
	}
}

func (table *Table) Quit() {
	table.quitting <- true
}
