package model

import (
	"fmt"
)

type Vec struct {
	X float32
	Y float32
}

func (v *Vec) String() string {
	return fmt.Sprintf("{%.2f, %.2f}", v.X, v.Y)
}
