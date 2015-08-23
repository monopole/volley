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
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/ifc"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/relay"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/options"
	"v.io/v23/rpc"
	_ "v.io/x/ref/runtime/factories/generic"
)

type vPlayer struct {
	p *model.Player
	c ifc.GameBuddyClientStub
}

type V23Manager struct {
	chatty               bool
	ctx                  *context.T
	shutdown             v23.Shutdown
	isRunning            bool
	leftDoor             model.DoorState
	rightDoor            model.DoorState
	rootName             string
	namespaceRoot        string
	rpcOpts              rpc.CallOpt
	relay                *relay.Relay
	myself               *model.Player
	players              []*vPlayer
	initialPlayerNumbers []int
	chBallCommand        <-chan model.BallCommand // Not owned, read from.
	chStop               chan chan bool           // Owned, read from.
	chNoNewBallsOrPeople chan chan bool           // Owned, read from.
	chDoorCommand        chan model.DoorCommand   // Owned, written to.
}

func NewV23Manager(
	chatty bool,
	rootName string,
	namespaceRoot string) *V23Manager {
	return &V23Manager{
		chatty,
		nil,          // ctx
		nil,          // shutdown
		false,        // isRunning
		model.Closed, // left door
		model.Closed, // right door
		rootName,
		namespaceRoot,
		options.SkipServerEndpointAuthorization{},
		nil, // relay
		nil, // myself
		[]*vPlayer{},
		nil,                  // initialPlayerNumbers
		nil,                  // chBallCommands
		make(chan chan bool), // chStop
		make(chan chan bool), // chNoNewBallsOrPeople
		make(chan model.DoorCommand),
	}
}

func gotNetwork() bool {
	_, err := http.Get(config.TestDomain)
	if err == nil {
		log.Printf("Network up - able to hit %s", config.TestDomain)
		return true
	}
	log.Printf("Something wrong with network: %v", err)
	return false
}

// Return true if ready to call Run
func (gm *V23Manager) IsReadyToRun() bool {
	if config.FailFast && !gotNetwork() {
		return false
	}
	if gm.chatty {
		log.Printf("Calling v23.Init")
	}
	gm.ctx, gm.shutdown = v23.Init()
	if gm.shutdown == nil {
		log.Panic("shutdown nil")
	}
	if gm.chatty {
		log.Printf("Setting root to %v", gm.namespaceRoot)
	}
	v23.GetNamespace(gm.ctx).SetRoots(gm.namespaceRoot)

	gm.initialPlayerNumbers = gm.playerNumbers()
	if gm.chatty {
		log.Printf("Found %d players.", len(gm.initialPlayerNumbers))
	}
	sort.Ints(gm.initialPlayerNumbers)
	myId := 1
	if len(gm.initialPlayerNumbers) > 0 {
		myId = gm.initialPlayerNumbers[len(gm.initialPlayerNumbers)-1] + 1
	}
	gm.myself = model.NewPlayer(myId)
	if gm.chatty {
		log.Printf("I am %v\n", gm.myself)
	}

	gm.relay = relay.MakeRelay()

	s := MakeServer(gm.ctx)
	myName := gm.serverName(gm.Me().Id())
	if gm.chatty {
		log.Printf("Calling myself %s\n", myName)
	}

	err := s.Serve(myName, ifc.GameBuddyServer(gm.relay), MakeAuthorizer())
	if err != nil {
		log.Panic("Error serving relay: ", err)
		return false
	}
	return true
}

func (gm *V23Manager) ChDoorCommand() <-chan model.DoorCommand {
	return gm.chDoorCommand
}

func (gm *V23Manager) ChIncomingBall() <-chan *model.Ball {
	if gm.relay != nil {
		return gm.relay.ChIncomingBall()
	}
	return nil
}

func (gm *V23Manager) Me() *model.Player {
	return gm.myself
}

func (gm *V23Manager) serverName(n int) string {
	return gm.rootName + fmt.Sprintf("%04d", n)
}

func (gm *V23Manager) recognizeOther(p *model.Player) {
	if gm.chatty {
		log.Printf("I (%v) am recognizing %v.", gm.Me(), p)
	}
	vp := &vPlayer{p, ifc.GameBuddyClient(gm.serverName(p.Id()))}

	// Keep the player list sorted.
	k := gm.findInsertion(p)
	gm.players = append(gm.players, nil)
	copy(gm.players[k+1:], gm.players[k:])
	gm.players[k] = vp

	if gm.chatty {
		log.Printf("I (%v) recognize %v.", gm.Me(), p)
	}
	if gm.isRunning {
		gm.checkDoors()
	} else {
		if gm.chatty {
			log.Printf("Not running, so not checking doors post recog.")
		}
	}
}

// Return index k of insertion point for the given player, given
// players sorted by Id.  The player currently at k-1 is on the 'left'
// of the argument, while the player at k is on the 'right'.  To
// insert, right-shift elements at k and above.
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

func (gm *V23Manager) forgetOther(p *model.Player) {
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
	gm.checkDoors()
}

func (gm *V23Manager) checkDoors() {
	if gm.chatty {
		log.Printf("Checking doors.\n")
	}
	if len(gm.players) == 0 {
		if gm.chatty {
			log.Printf("I'm the only player.")
		}
		gm.assureDoor(model.DoorCommand{model.Closed, model.Left})
		gm.assureDoor(model.DoorCommand{model.Closed, model.Right})
	} else if gm.myself.Id() < gm.players[0].p.Id() {
		if gm.chatty {
			log.Printf("I'm the left-most of %d players.\n", len(gm.players)+1)
		}
		gm.assureDoor(model.DoorCommand{model.Closed, model.Left})
		gm.assureDoor(model.DoorCommand{model.Open, model.Right})
	} else if gm.players[len(gm.players)-1].p.Id() < gm.myself.Id() {
		if gm.chatty {
			log.Printf("I'm the right-most of %d players.\n", len(gm.players)+1)
		}
		gm.assureDoor(model.DoorCommand{model.Open, model.Left})
		gm.assureDoor(model.DoorCommand{model.Closed, model.Right})
	} else {
		if gm.chatty {
			log.Printf("I'm somewhere in the middle.\n")
		}
		gm.assureDoor(model.DoorCommand{model.Open, model.Left})
		gm.assureDoor(model.DoorCommand{model.Open, model.Right})
	}
	if gm.chatty {
		log.Println("Current players: ", gm.playersString())
	}
}

func (gm *V23Manager) playersString() (s string) {
	k := gm.findInsertion(gm.myself)
	s = ""
	for i := 0; i < k; i++ {
		s += gm.players[i].p.String() + " "
	}
	if gm.leftDoor == model.Open {
		s += "_"
	} else {
		s += "["
	}
	s += gm.myself.String()
	if gm.rightDoor == model.Open {
		s += "_"
	} else {
		s += "]"
	}
	s += " "
	for i := k; i < len(gm.players); i++ {
		s += gm.players[i].p.String() + " "
	}
	return
}

func (gm *V23Manager) assureDoor(dc model.DoorCommand) {
	switch dc {
	case model.DoorCommand{model.Open, model.Left}:
		if gm.leftDoor == model.Open {
			if gm.chatty {
				log.Printf("Left door already open.\n")
			}
			return
		}
		gm.leftDoor = model.Open
	case model.DoorCommand{model.Open, model.Right}:
		if gm.rightDoor == model.Open {
			if gm.chatty {
				log.Printf("Right door already open.\n")
			}
			return
		}
		gm.rightDoor = model.Open
	case model.DoorCommand{model.Closed, model.Left}:
		if gm.leftDoor == model.Closed {
			if gm.chatty {
				log.Printf("Left door already closed.\n")
			}
			return
		}
		gm.leftDoor = model.Closed
	case model.DoorCommand{model.Closed, model.Right}:
		if gm.rightDoor == model.Closed {
			if gm.chatty {
				log.Printf("Right door already closed.\n")
			}
			return
		}
		gm.rightDoor = model.Closed
	}
	if gm.chDoorCommand == nil {
		log.Panic("The door channel is nil.")
	}
	if gm.chatty {
		log.Printf("Sending door command: %v\n", dc)
	}
	gm.chDoorCommand <- dc
	if gm.chatty {
		log.Printf("Door command %v consumed.\n", dc)
	}
}

func (gm *V23Manager) sayHelloToEveryone() {
	if gm.chatty {
		log.Printf("Me (%v) saying Hello to %d other players.\n",
			gm.Me(), len(gm.players))
	}
	wp := ifc.Player{int32(gm.Me().Id())}
	for _, vp := range gm.players {
		if gm.chatty {
			log.Printf("RPC sending: asking %v to recognize me=%v", vp, gm.Me())
			log.Printf("  gm.ctx %T = %v", gm.ctx, gm.ctx)
			log.Printf("  wp %T = %v", wp, wp)
		}
		if err := vp.c.Recognize(gm.ctx, wp, gm.rpcOpts); err != nil {
			// TODO: Instead of panicing, just drop the player from the players list.
			log.Panic("Recognize failed: ", err)
		}
		if gm.chatty {
			log.Printf("RPC Recognize call completed!")
		}
	}
	if gm.chatty {
		log.Printf("Me (%v) DONE saying Hello.\n", gm.Me())
	}
}

func (gm *V23Manager) sayGoodbyeToEveryone() {
	if gm.chatty {
		log.Println("Saying goodbye to other players.")
	}
	wp := ifc.Player{int32(gm.Me().Id())}
	for _, vp := range gm.players {
		if gm.chatty {
			log.Printf("RPC sending: asking %v to forget me=%v", vp.p, gm.Me())
			log.Printf("  gm.ctx %T = %v", gm.ctx, gm.ctx)
			log.Printf("  wp %T = %v", wp, wp)
		}
		if err := vp.c.Forget(gm.ctx, wp, gm.rpcOpts); err != nil {
			log.Println("Forget failed, but continuing; err=", err)
		}
		if gm.chatty {
			log.Println("Forget call completed.")
		}
	}
}

// Return array of known players.
func (gm *V23Manager) playerNumbers() (list []int) {
	list = []int{}
	rCtx, cancel := context.WithTimeout(gm.ctx, time.Minute)
	defer cancel()
	if gm.chatty {
		log.Printf("Recovering namespace.")
	}
	ns := v23.GetNamespace(rCtx)
	if gm.chatty {
		log.Printf("namespace == %T %v", ns, ns)
	}
	pattern := gm.rootName + "*"
	if gm.chatty {
		log.Printf("Calling glob with %T=%v, pattern=%v\n", rCtx, rCtx, pattern)
	}
	c, err := ns.Glob(rCtx, pattern)
	if err != nil {
		log.Printf("ns.Glob(%v) failed: %v", pattern, err)
		return
	}
	if gm.chatty {
		log.Printf("Awaiting response from Glob request.")
	}
	for res := range c {
		if gm.chatty {
			log.Printf("Got a result: %v\n", res)
		}
		switch v := res.(type) {
		case *naming.GlobReplyEntry:
			name := v.Value.Name
			if gm.chatty {
				log.Printf("Raw name is: %v\n", name)
			}
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
	if gm.chatty {
		log.Printf("Glob result channel exhausted.")
	}
	return
}

func (gm *V23Manager) Run(cbc <-chan model.BallCommand) {
	if gm.chatty {
		log.Println("Final prep of V23Manager.")
	}
	gm.chBallCommand = cbc
	for _, id := range gm.initialPlayerNumbers {
		gm.recognizeOther(model.NewPlayer(id))
	}
	gm.sayHelloToEveryone()
	gm.isRunning = true
	gm.checkDoors()
	if gm.chatty {
		log.Println("Starting V23Manager run loop.")
	}
	for {
		select {
		case ch := <-gm.chStop:
			gm.stop()
			ch <- true
			return
		case ch := <-gm.chNoNewBallsOrPeople:
			gm.noNewBallsOrPeople()
			ch <- true
		case bc := <-gm.chBallCommand:
			gm.throwBall(bc)
		case p := <-gm.relay.ChRecognize():
			gm.recognizeOther(p)
		case p := <-gm.relay.ChForget():
			gm.forgetOther(p)
		}
	}
}

func (gm *V23Manager) throwBall(bc model.BallCommand) {
	if gm.chatty {
		log.Printf("v23 manager got ball throw command: %v\n", bc)
	}
	k := gm.findInsertion(gm.myself)
	if bc.D == model.Left {
		// Throw ball left.
		k--
		if k >= 0 {
			gm.sendBallRpc(bc, gm.players[k])
		} else {
			log.Panic("Nobody on left!  Send back to table.")
		}
	} else {
		// Throw ball right.
		if k <= len(gm.players)-1 {
			gm.sendBallRpc(bc, gm.players[k])
		} else {
			log.Panic("Nobody on right!  Send back to table.")
		}
	}
}

func (gm *V23Manager) sendBallRpc(bc model.BallCommand, vp *vPlayer) {
	wb := serializeBall(bc.B)
	log.Printf("RPC sending: throwing ball %v to %v : %v\n", bc.D, vp.p, vp.c)
	if err := vp.c.Accept(gm.ctx, wb, gm.rpcOpts); err != nil {
		log.Panic("Ball throw %v failed; err=%v", bc.D, err)
	}
	log.Printf("Ball throw %v RPC done.", bc.D)
}

func serializeBall(b *model.Ball) ifc.Ball {
	wp := ifc.Player{int32(b.Owner().Id())}
	return ifc.Ball{
		wp, b.GetPos().X, b.GetPos().Y, b.GetVel().X, b.GetVel().Y}
}

func (gm *V23Manager) NoNewBallsOrPeople() {
	ch := make(chan bool)
	gm.chNoNewBallsOrPeople <- ch
	<-ch
}

func (gm *V23Manager) noNewBallsOrPeople() {
	if gm.chatty {
		log.Println("********************* No New Balls or people.")
	}
	gm.relay.StopAcceptingPlayers()
	gm.sayGoodbyeToEveryone()
	gm.relay.StopAcceptingBalls()
}

func (gm *V23Manager) Stop() {
	ch := make(chan bool)
	gm.chStop <- ch
	<-ch
}

func (gm *V23Manager) stop() {
	if gm.chatty {
		log.Println("v23 calling native shutdown.")
	}
	gm.shutdown()
	gm.relay.Stop()
	if gm.chatty {
		log.Println("v23: closing door command channel.")
	}
	close(gm.chDoorCommand)
	if gm.chatty {
		log.Println("v23 runtime done.")
	}
}
