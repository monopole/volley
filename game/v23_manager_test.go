package game

import (
	"log"
	"testing"
	"time"
)

func TestSomething(t *testing.T) {
	log.Println("start test")
	chChan := make(chan chan bool)
	log.Println("making mgr")
	m := NewV23Manager(nil, chChan)
	log.Println("got manager")
	//go func() {
	//	log.Println("start timer")
	//	<-time.After(3 * time.Second)
	//		log.Println("done timer")
	//		ch <- true
	// }()
	log.Println("start run")
	go m.run()
	log.Println("start timer")
	<-time.After(3 * time.Second)
	log.Println("done timer")
	chBool := make(chan bool)
	log.Println("foo1")
	chChan <- chBool
	log.Println("foo2")
	<-chBool
	log.Println("foo3")
}
