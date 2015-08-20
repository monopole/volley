package model

import (
	"fmt"
)

type Ball struct {
	owner *Player
	p     Vec
	v     Vec
}

func NewBall(
	owner *Player,
	p Vec, v Vec) *Ball {
	return &Ball{owner, p, v}
}

func (b *Ball) String() string {
	owner := "zod"
	if b.owner != nil {
		owner = b.owner.String()
	}
	return fmt.Sprintf(
		"(%s p%s v%s)", owner, b.p.String(), b.v.String())
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
