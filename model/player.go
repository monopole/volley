package model

import (
	"strconv"
)

type Player struct {
	id int
}

func (b *Player) String() string {
	return strconv.Itoa(b.id)
}

func (b *Player) Id() int {
	return b.id
}

func NewPlayer(id int) *Player {
	return &Player{id}
}
