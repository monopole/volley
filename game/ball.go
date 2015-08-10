package game

// Ball state and operations.

// TRUE?? Touch events report an x and y in normalized units where the center
// of the screen is (0, 0), the left side of the screen is (-1,y), the
// top is (x, 1), etc.
// To place something ...

import (
	"fmt"
)

type Ball struct {
	id int
	x  float32
	y  float32
	dx float32
	dy float32
}

func NewBall(id int) *Ball {
	return &Ball{id, 0, 0, 0, 0}
}

func (b *Ball) String() string {
	return fmt.Sprintf("(%d {%v, %v} d{%v, %v})", b.id, b.x, b.y, b.dx, b.dy)
}

func (b *Ball) Id() int {
	return b.id
}

func (b *Ball) GetPos() (float32, float32) {
	return b.x, b.y
}

func (b *Ball) GetVelocity() (float32, float32) {
	return b.dx, b.dy
}

func (b *Ball) SetPos(ix float32, iy float32) {
	b.x = ix
	b.y = iy
}
