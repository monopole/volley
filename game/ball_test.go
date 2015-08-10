package game

import (
	"log"
	"testing"
)

func TestSomething(t *testing.T) {
	b := NewBall(3)
	log.Println(b.Id())
}
