package model

import (
	"log"
	"testing"
)

func TestSomething(t *testing.T) {
	b := NewBall(NewPlayer(3), 1, 2, 3, 4)
	log.Println(b)
}
