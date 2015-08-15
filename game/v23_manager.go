// V23Manager is a peer to other instances of same on the net.
//
// Each device/game/program instance must have one V23Manager.
//
// Each has an embedded V23 service, and is a direct client to the V23
// services held by all the other instances.
//
// On startup, the manager finds all the other instances via a
// mounttable, figures out what it should call itself, and fires off
// go routines to manage data coming in on various channels, and
// establishes contact with the other players.

package game

import (
	"fmt"
	"github.com/monopole/croupier/ifc"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/service"
	"log"
	"sort"
	"strconv"
	"time"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/options"
	//	"v.io/v23/options"
	_ "v.io/x/ref/runtime/factories/generic"
)

type vPlayer struct {
	p *model.Player
	c ifc.GameBuddyClientStub
}

type V23Manager struct {
	ctx                  *context.T
	shutdown             v23.Shutdown
	rootName             string
	namespaceRoot        string
	relay                *service.Relay
	chatty               bool
	myself               *model.Player
	players              []*vPlayer
	initialPlayerNumbers []int
	chBallCommand        <-chan model.BallCommand // Not owned, read from.
	chQuit               chan chan bool           // Owned, read from.
	chDoorCommand        chan model.DoorCommand   // Owned, written to.
}

func NewV23Manager(rootName string, namespaceRoot string) *V23Manager {
	ctx, shutdown := v23.Init()
	if shutdown == nil {
		log.Panic("shutdown nil")
	}
	return &V23Manager{
		ctx, shutdown,
		rootName, namespaceRoot,
		nil,  // relay
		true, // chatty
		nil,  // myself
		[]*vPlayer{},
		nil, // initialPlayerNumbers
		nil, // chBallCommands
		make(chan chan bool),
		make(chan model.DoorCommand),
	}
}

func (gm *V23Manager) Initialize() {
	if gm.chatty {
		log.Printf("Scanning namespace %v\n", gm.namespaceRoot)
	}
	v23.GetNamespace(gm.ctx).SetRoots(gm.namespaceRoot)

	gm.initialPlayerNumbers = gm.playerNumbers()
	sort.Ints(gm.initialPlayerNumbers)
	myId := 1
	if len(gm.initialPlayerNumbers) > 0 {
		myId = gm.initialPlayerNumbers[len(gm.initialPlayerNumbers)-1] + 1
	}
	gm.myself = model.NewPlayer(myId)
	if gm.chatty {
		log.Printf("I am %v\n", gm.myself)
	}

	gm.relay = service.MakeRelay()

	s := MakeServer(gm.ctx)
	myName := gm.serverName(gm.Me().Id())
	if gm.chatty {
		log.Printf("Calling myself %s\n", myName)
	}

	err := s.Serve(myName, ifc.GameBuddyServer(gm.relay), MakeAuthorizer())
	if err != nil {
		log.Panic("Error serving service: ", err)
	}
}

func (gm *V23Manager) ChDoorCommand() <-chan model.DoorCommand {
	return gm.chDoorCommand
}

func (gm *V23Manager) ChQuit() chan<- chan bool {
	return gm.chQuit
}

func (gm *V23Manager) ChIncomingBall() <-chan *model.Ball {
	return gm.relay.ChBall()
}

func (gm *V23Manager) Me() *model.Player {
	return gm.myself
}

func (gm *V23Manager) serverName(n int) string {
	return gm.rootName + fmt.Sprintf("%04d", n)
}

func (gm *V23Manager) recognize(p *model.Player) {
	if gm.chatty {
		log.Printf("%v recognizing %v.", gm.Me(), p)
	}
	vp := &vPlayer{p, ifc.GameBuddyClient(gm.serverName(p.Id()))}

	// Keep the player list sorted.
	k := gm.findInsertion(p)
	gm.players = append(gm.players, nil)
	copy(gm.players[k+1:], gm.players[k:])
	gm.players[k] = vp

	// Someone came in on the left,
	command := model.DoorCommand{model.Open, model.Left}
	if p.Id() > gm.myself.Id() {
		// else someone came in on the right.
		command = model.DoorCommand{model.Open, model.Right}
	}
	if gm.chatty {
		log.Printf("Sending a door open command.\n")
		if gm.chDoorCommand == nil {
			log.Panic("The channel is nil.")
		}
	}
	// go func() {
	gm.chDoorCommand <- command
	// }()
	if gm.chatty {
		log.Printf("Door is apparently open.\n")
	}
}

func (gm *V23Manager) findInsertion(p *model.Player) int {
	for k, member := range gm.players {
		if p.Id() < member.p.Id() {
			return k
		}
	}
	return len(gm.players)
}

func (gm *V23Manager) findPlayerIndex(p *model.Player) int {
	return findIndex(len(gm.players),
		func(i int) bool { return gm.players[i].p.Id() == p.Id() })
}

func findIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func (gm *V23Manager) forget(p *model.Player) {
	i := gm.findPlayerIndex(p)
	if i > -1 {
		if gm.chatty {
			log.Printf("Me=(%v) forgetting %v.\n", gm.Me(), p)
		}
		gm.players = append(gm.players[:i], gm.players[i+1:]...)
	} else {
		if gm.chatty {
			log.Printf("Asked to forget %v, but don't know him\n.", p)
		}
	}
	if len(gm.players) == 0 {
		if gm.chatty {
			log.Printf("I'm the only player.\n")
		}
		// go func() {
		gm.chDoorCommand <- model.DoorCommand{model.Closed, model.Left}
		gm.chDoorCommand <- model.DoorCommand{model.Closed, model.Right}
		// }()
	} else if gm.players[0].p.Id() == gm.myself.Id() {
		if gm.chatty {
			log.Printf("Sending a door close left command.\n")
		}
		// go func() {
		gm.chDoorCommand <- model.DoorCommand{model.Closed, model.Left}
		// }()
	} else if gm.players[len(gm.players)-1].p.Id() == gm.myself.Id() {
		if gm.chatty {
			log.Printf("Sending a door close right command.\n")
		}
		// go func() {
		gm.chDoorCommand <- model.DoorCommand{model.Closed, model.Right}
		// }()
	}
}

func (gm *V23Manager) sayHelloToEveryone() {
	wp := ifc.Player{int32(gm.Me().Id())}
	for _, vp := range gm.players {
		if gm.chatty {
			log.Printf("Asking %v to recognize me=%v\n", vp.p, gm.Me())
		}
		if err := vp.c.Recognize(
			gm.ctx, wp, options.SkipServerEndpointAuthorization{}); err != nil {
			log.Panic("Recognize failed: ", err)
		}
		if gm.chatty {
			log.Printf("Recognize call completed.")
		}
	}
}

func (gm *V23Manager) sayGoodbyeToEveryone() {
	if gm.chatty {
		log.Println("Saying goodbye to other players.")
	}
	wp := ifc.Player{int32(gm.Me().Id())}
	for _, vp := range gm.players {
		if gm.chatty {
			log.Printf("Asking %v to forget me=%v\n", vp.p, gm.Me())
		}
		if err := vp.c.Forget(
			gm.ctx, wp, options.SkipServerEndpointAuthorization{}); err != nil {
			log.Panic("Forget failed: ", err)
		}
		if gm.chatty {
			log.Printf("Forget call completed.")
		}
	}
}

// Return array of known players.
func (gm *V23Manager) playerNumbers() (list []int) {
	list = []int{}
	rCtx, cancel := context.WithTimeout(gm.ctx, time.Minute)
	defer cancel()
	ns := v23.GetNamespace(rCtx)
	pattern := gm.rootName + "*"
	c, err := ns.Glob(rCtx, pattern)
	if err != nil {
		log.Printf("ns.Glob(%q) failed: %v", pattern, err)
		return
	}
	for res := range c {
		switch v := res.(type) {
		case *naming.GlobReplyEntry:
			name := v.Value.Name
			if name != "" {
				putativeNumber := name[len(gm.rootName):]
				n, err := strconv.ParseInt(putativeNumber, 10, 32)
				if err != nil {
					log.Println(err)
				} else {
					list = append(list, int(n))
				}
				if gm.chatty {
					log.Println("Found player: ", v.Value.Name)
				}
			}
		default:
		}
	}
	return
}

func (gm *V23Manager) Run(cbc <-chan model.BallCommand) {
	gm.chBallCommand = cbc
	if gm.chatty {
		log.Println("V23Manager Running.")
	}
	for _, id := range gm.initialPlayerNumbers {
		gm.recognize(model.NewPlayer(id))
	}
	gm.sayHelloToEveryone()
	for {
		select {
		case ch := <-gm.chQuit:
			gm.quit()
			ch <- true
			return
		case bc := <-gm.chBallCommand:
			gm.tossBall(bc)
		case p := <-gm.relay.ChRecognize():
			gm.recognize(p)
		case p := <-gm.relay.ChForget():
			gm.forget(p)
		}
	}
}

func serializeBall(b *model.Ball) ifc.Ball {
	wp := ifc.Player{int32(b.Owner().Id())}
	return ifc.Ball{
		wp, b.GetPos().X, b.GetPos().Y, b.GetVel().X, b.GetVel().Y}
}

func (gm *V23Manager) tossBall(bc model.BallCommand) {
	if gm.chatty {
		log.Printf("Got ball command: %v\n", bc)
	}

	k := gm.findInsertion(gm.myself)
	wb := serializeBall(bc.B)

	if bc.D == model.Left {
		// Throw ball left.
		k--
		if k >= 0 {
			vp := gm.players[k]
			if err := vp.c.Accept(
				gm.ctx, wb, options.SkipServerEndpointAuthorization{}); err != nil {
				log.Panic("Ball throw left failed.")
			}
		} else {
			// Nobody on the left!
			// Send ball back into table - at the moment, don't have the channel.
			log.Panic("1 Refactor to get ball channel.")
		}
	} else {
		// Throw ball right.
		if k <= len(gm.players)-1 {
			vp := gm.players[k]
			if err := vp.c.Accept(
				gm.ctx, wb, options.SkipServerEndpointAuthorization{}); err != nil {
				log.Panic("Ball throw right failed.")
			}
		} else {
			// Nobody on the right!
			// Send ball back into table - at the moment, don't have the channel.
			log.Panic("2 Refactor to get ball channel.")
		}
	}
}

func (gm *V23Manager) quit() {
	gm.sayGoodbyeToEveryone()
	if gm.chatty {
		log.Println("Shutting down v23 runtime.")
	}
	gm.shutdown()
	if gm.chatty {
		log.Println("v23 runttime done.")
	}
	gm.relay.Close()
	if gm.chatty {
		log.Println("Relay closed.")
	}
	close(gm.chDoorCommand)
}
