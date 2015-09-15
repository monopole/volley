package main

import (
	"fmt"
	"github.com/monopole/volley/config"
	"github.com/monopole/volley/net"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need args")
		return
	}
	nsRoot := "/" + net.DetermineNamespaceRoot()
	log.Printf("Using v23.namespace.root=%s", nsRoot)
	nm := net.NewV23Manager(
		config.Chatty, config.RootName, true, nsRoot)

	chReady := nm.GetReady()

	select {
	case <-time.After(5 * time.Second):
		log.Printf("Ready loop timed out.\n")
		return
	case ready := <-chReady:
		if !ready {
			log.Printf("Seem unable to start NM.\n")
			return
		}
	}
	nm.JoinGame(nil)
	if config.Chatty {
		log.Printf("NM now running.\n")
	}

	switch os.Args[1] {
	case "list":
		nm.List()
	case "mc":
		if len(os.Args[2]) > 0 {
			nm.DoMasterCommand(os.Args[2])
		} else {
			log.Println("Don't understand mc arg")
		}
	case "kick":
		nm.Kick()
	case "quit":
		id, _ := strconv.Atoi(os.Args[2])
		nm.Quit(id)
	case "fire":
		count, _ := strconv.Atoi(os.Args[2])
		nm.FireBall(count)
	case "pause":
		x, _ := strconv.ParseFloat(os.Args[2], 32)
		pd := float32(x)
		nm.SetPauseDuration(pd)
	case "gravity":
		x, _ := strconv.ParseFloat(os.Args[2], 32)
		g := float32(x)
		nm.SetGravity(g)
	default:
		log.Printf("Don't understand: %s\n", os.Args[1])
	}
}
