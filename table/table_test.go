package table

import (
	"log"
	"testing"
	"time"
)

func makeTable(isDone chan bool) *Table {
	return NewTable(4,
		isDone,
		make(chan commandType),
		make(chan *Player),
		make(chan int))
}

func TestTableMake(t *testing.T) {
	log.Println(makeTable(make(chan bool)))
}

func TestTableQuit(t *testing.T) {
	isDone := make(chan bool)
	table := makeTable(isDone)
	go table.play()
	go func() {
		time.Sleep(1 * time.Second)
		table.Quit()
	}()
	<-isDone
}
