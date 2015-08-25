package main

import (
	"fmt"
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/game"
	"log"
	"os"
	"strconv"
)

func main() {
	fmt.Println(os.Args)
	if len(os.Args) < 1 {
		return
	}

	gm := game.NewV23Manager(
		config.Chatty, config.RootName, config.NamespaceRoot)
	if !gm.IsReadyToRun(true) {
		if config.Chatty {
			log.Printf("gm not ready!")
		}
		return
	}
	gm.RunPrep(nil)

	if os.Args[1] == "list" {
		log.Printf("listing!")
		gm.List()
		return
	}
	if os.Args[1] == "kick" {
		log.Printf("kicking!")
		gm.Kick()
		return
	}
	if os.Args[1] == "quit" {
		id, _ := strconv.Atoi(os.Args[2])
		log.Printf("quitting %d", id)
		gm.Quit(id)
		return
	}
	log.Printf("Don't understand: %s\n", os.Args[1])
}
