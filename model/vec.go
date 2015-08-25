package model

import (
	"fmt"
)

type Vec struct {
	X float32
	Y float32
}

func (v *Vec) String() string {
	return fmt.Sprintf("{%.4f, %.4f}", v.X, v.Y)
}
