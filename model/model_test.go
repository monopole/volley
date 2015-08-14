package model

import (
	"testing"
)

func TestSomething(t *testing.T) {
	expect := "(3 p{1 2} v{3 4})"
	b := NewBall(NewPlayer(3), Vec{1, 2}, Vec{3, 4})
	if b.String() != expect {
		t.Errorf("Expected \"%s\", got \"%s\"", expect, b.String())
	}
}
