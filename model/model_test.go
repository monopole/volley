package model

import (
	"log"
	"testing"
)

func TestSomething(t *testing.T) {
	b := NewBall(NewPlayer(3))
	log.Println(b)
}
