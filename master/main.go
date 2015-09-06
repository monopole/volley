package main

import (
	"fmt"
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/net"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need args")
		return
	}
	nsRoot := "/" + net.DetermineNamespaceRoot()
	log.Printf("Using v23.namespace.root=%s", nsRoot)
	gm := net.NewV23Manager(
		config.Chatty, config.RootName, nsRoot)
	if !gm.IsReadyToRun(true) {
		if config.Chatty {
			log.Printf("gm not ready!")
		}
		return
	}
	gm.RunPrep(nil)
	switch os.Args[1] {
	case "list":
		gm.List()
	case "mc":
		if len(os.Args[2]) > 0 {
			gm.DoMasterCommand(os.Args[2])
		} else {
			log.Println("Don't understand mc arg")
		}
	case "kick":
		gm.Kick()
	case "quit":
		id, _ := strconv.Atoi(os.Args[2])
		gm.Quit(id)
	case "fire":
		count, _ := strconv.Atoi(os.Args[2])
		gm.FireBall(count)
	case "pause":
		x, _ := strconv.ParseFloat(os.Args[2], 32)
		pd := float32(x)
		gm.SetPauseDuration(pd)
	case "gravity":
		x, _ := strconv.ParseFloat(os.Args[2], 32)
		g := float32(x)
		gm.SetGravity(g)
	default:
		log.Printf("Don't understand: %s\n", os.Args[1])
	}
}
