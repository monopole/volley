package model

import (
	"fmt"
	"strconv"
)

type ExecCommand int

const (
	ExecError ExecCommand = iota
	ExecStart
	ExecStop
	ExecRandomImpulse
	ExecPaint
)

type Vec struct {
	X float32
	Y float32
}

func (v *Vec) String() string {
	return fmt.Sprintf("{%v, %v}", v.X, v.Y)
}

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

const (
	UnknownPlayerId = 0
)

type Ball struct {
	owner *Player
	p     Vec
	v     Vec
}

type DoorState int

const (
	Open DoorState = iota
	Closed
)

type Direction int

const (
	Left Direction = iota
	Right
)

type DoorCommand struct {
	S DoorState
	D Direction
}

func (dc DoorCommand) String() string {
	if dc.S == Open {
		if dc.D == Left {
			return "open-left"
		} else {
			return "open-right"
		}
	} else {
		if dc.D == Left {
			return "close-left"
		} else {
			return "close-right"
		}
	}
}

type BallCommand struct {
	B *Ball
	D Direction
}

func (bc BallCommand) String() string {
	if bc.D == Left {
		return "toss-" + bc.B.String() + "-left"
	}
	return "toss-" + bc.B.String() + "-right"
}

func NewBall(
	owner *Player,
	p Vec, v Vec) *Ball {
	return &Ball{owner, p, v}
}

func (b *Ball) String() string {
	return fmt.Sprintf(
		"(%v p%v v%v)",
		b.owner, b.p, b.v)
}

func (b *Ball) Owner() *Player {
	return b.owner
}

func (b *Ball) GetPos() Vec {
	return b.p
}

func (b *Ball) GetVel() Vec {
	return b.v
}

func (b *Ball) SetPos(x float32, y float32) {
	b.p = Vec{x, y}
}

func (b *Ball) SetVel(x float32, y float32) {
	b.v = Vec{x, y}
}

// Interface to other game players on the net.
//
// An implementation of this interface will initialize a runtime
// capable of making RPCs and begin to accept and send RPCs on behalf
// of the client of the interface.
type Game interface {
	// Returns a channel on which new players may appear.
	Players() chan Player

	// Send a ball to some other player.  The target player is decided
	// by the table, since the table knows how the ball is moving and
	// how the players are arranged.
	SendBall(*Ball, *Player) error

	// Tell all players that you are leaving, and won't be sending or
	// accepting any more data.
	Quit() error
}
