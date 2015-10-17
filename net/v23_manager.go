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

package net

import (
	"fmt"
	"github.com/monopole/volley/config"
	"github.com/monopole/volley/ifc"
	"github.com/monopole/volley/model"
	"github.com/monopole/volley/relay"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"sync"

	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/options"
	"v.io/v23/rpc"
	"v.io/v23/security"
	_ "v.io/x/ref/runtime/factories/generic"
)

const (
	useFixedNs = true
	theFixedNs = "192.168.86.254:8101"
)

type vPlayer struct {
	p *model.Player
	c ifc.GameServiceClientStub
}

type V23Manager struct {
	chatty               bool
	ctx                  *context.T
	shutdown             v23.Shutdown
	isRunning            bool
	isGameMaster         bool
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
	mu                   *sync.RWMutex
	isReady              bool
}

func NewV23Manager(
	chatty bool,
	rootName string,
	isGameMaster bool,
	namespaceRoot string) *V23Manager {
	return &V23Manager{
		chatty,
		nil,          // ctx
		nil,          // shutdown
		false,        // isRunning
		isGameMaster, // isGameMaster
		model.Closed, // left door
		model.Closed, // right door
		rootName,
		namespaceRoot,
		options.ServerAuthorizer{security.AllowEveryone()},
		nil, // relay
		nil, // myself
		[]*vPlayer{},
		nil,                  // initialPlayerNumbers
		nil,                  // chBallCommands
		make(chan chan bool), // chStop
		make(chan chan bool), // chNoNewBallsOrPeople
		make(chan model.DoorCommand),
		new(sync.RWMutex),
		false,
	}
}

var reNsRoot *regexp.Regexp

func init() {
	reNsRoot, _ = regexp.Compile("v23\\.namespace\\.root=([a-z\\.0-9:]+)")
}

func DetermineNamespaceRoot() string {
	if useFixedNs {
		return theFixedNs
	}

	res, err := http.Get(config.TestDomain)
	if err != nil {
		log.Printf("Unable to Get %s", config.TestDomain)
		return config.NamespaceRoot
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("Problem grabbing content from %s", config.TestDomain)
		return config.NamespaceRoot
	}
	chuckles := reNsRoot.FindStringSubmatch(string(content))
	if len(chuckles) > 1 {
		return chuckles[1]
	}
	log.Printf("Got web text, but unable to parse using %s", reNsRoot)
	return config.NamespaceRoot
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

func (nm *V23Manager) IsRunning() bool {
	return nm.isRunning
}

// GetReady returns a bool channel.  The channel will get one datum
// during it's life.  If the datum is true, the manager is ready to
// join the game.  If the datum is false, the manager will never be
// ready within the contraints of its own timeouts.  The client can
// call this multiple times, but parallel calls block till this
// finishes.
func (nm *V23Manager) GetReady() <-chan bool {
	nm.mu.Lock()
	ch := make(chan bool)
	if nm.isReady {
		go func() {
			ch <- true
		}()
		nm.mu.Unlock()
		return ch
	}
	if config.FailFast && !gotNetwork() {
		go func() {
			ch <- false
		}()
		nm.mu.Unlock()
		return ch
	}
	go nm.getReadyToRun(ch)
	return ch
}

func (nm *V23Manager) getReadyToRun(ch chan bool) {
	defer nm.mu.Unlock()
	if nm.chatty {
		log.Printf("Calling v23.Init")
	}
	nm.ctx, nm.shutdown = v23.Init()
	if nm.shutdown == nil {
		log.Panic("shutdown nil")
	}
	if nm.chatty {
		log.Printf("Setting root to %v", nm.namespaceRoot)
	}
	v23.GetNamespace(nm.ctx).SetRoots(nm.namespaceRoot)

	nm.initialPlayerNumbers = nm.playerNumbers()
	if nm.chatty {
		log.Printf("Found %d players.", len(nm.initialPlayerNumbers))
	}
	sort.Ints(nm.initialPlayerNumbers)
	myId := 1
	if len(nm.initialPlayerNumbers) > 0 {
		myId = nm.initialPlayerNumbers[len(nm.initialPlayerNumbers)-1] + 1
	}

	if nm.isGameMaster {
		myId = 999
	}

	nm.relay = relay.MakeRelay()
	nm.myself = model.NewPlayer(myId)
	if nm.isGameMaster {
		if nm.chatty {
			log.Printf("I am game master.")
		}
		nm.isReady = true
		ch <- true
		return
	}
	if nm.chatty {
		log.Printf("I am player %v\n", nm.myself)
	}

	myName := nm.serverName(nm.Me().Id())
	if nm.chatty {
		log.Printf("Calling myself %s\n", myName)
	}
	ctx, s, err := v23.WithNewServer(nm.ctx, myName, ifc.GameServiceServer(nm.relay), MakeAuthorizer())
	if err != nil {
		log.Panic("Error creating server:", err)
		ch <- false
		return
	}
	saveEndpointToFile(s)
	nm.ctx = ctx
	nm.isReady = true
	ch <- true
}

func (nm *V23Manager) ChDoorCommand() <-chan model.DoorCommand {
	return nm.chDoorCommand
}

func (nm *V23Manager) GetRelay() model.Relay {
	return nm.relay
}

func (nm *V23Manager) Me() *model.Player {
	return nm.myself
}

func (nm *V23Manager) serverName(n int) string {
	return nm.rootName + fmt.Sprintf("%04d", n)
}

func (nm *V23Manager) recognizeOther(p *model.Player) {
	if nm.chatty {
		log.Printf("I (%v) am recognizing %v.", nm.Me(), p)
	}
	vp := &vPlayer{p, ifc.GameServiceClient(nm.serverName(p.Id()))}

	// Keep the player list sorted.
	k := nm.findInsertion(p)
	nm.players = append(nm.players, nil)
	copy(nm.players[k+1:], nm.players[k:])
	nm.players[k] = vp

	if nm.chatty {
		log.Printf("I (%v) recognize %v.", nm.Me(), p)
	}
	if nm.isRunning {
		nm.checkDoors()
	} else {
		if nm.chatty {
			log.Printf("Not running, so not checking doors post recog.")
		}
	}
}

// Return index k of insertion point for the given player, given
// players sorted by Id.  The player currently at k-1 is on the 'left'
// of the argument, while the player at k is on the 'right'.  To
// insert, right-shift elements at k and above.
func (nm *V23Manager) findInsertion(p *model.Player) int {
	for k, member := range nm.players {
		if p.Id() < member.p.Id() {
			return k
		}
	}
	return len(nm.players)
}

func (nm *V23Manager) findPlayerIndex(p *model.Player) int {
	return findIndex(len(nm.players),
		func(i int) bool { return nm.players[i].p.Id() == p.Id() })
}

func findIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func (nm *V23Manager) forgetOther(p *model.Player) {
	i := nm.findPlayerIndex(p)
	if i > -1 {
		if nm.chatty {
			log.Printf("Me=(%v) forgetting %v.\n", nm.Me(), p)
		}
		nm.players = append(nm.players[:i], nm.players[i+1:]...)
	} else {
		if nm.chatty {
			log.Printf("Asked to forget %v, but don't know him\n.", p)
		}
	}
	nm.checkDoors()
}

func (nm *V23Manager) checkDoors() {
	if nm.chatty {
		log.Printf("Checking doors.\n")
	}
	if len(nm.players) == 0 {
		if nm.chatty {
			log.Printf("I'm the only player.")
		}
		nm.assureDoor(model.DoorCommand{model.Closed, model.Left})
		nm.assureDoor(model.DoorCommand{model.Closed, model.Right})
	} else if nm.myself.Id() < nm.players[0].p.Id() {
		if nm.chatty {
			log.Printf("I'm the left-most of %d players.\n", len(nm.players)+1)
		}
		nm.assureDoor(model.DoorCommand{model.Closed, model.Left})
		nm.assureDoor(model.DoorCommand{model.Open, model.Right})
	} else if nm.players[len(nm.players)-1].p.Id() < nm.myself.Id() {
		if nm.chatty {
			log.Printf("I'm the right-most of %d players.\n", len(nm.players)+1)
		}
		nm.assureDoor(model.DoorCommand{model.Open, model.Left})
		nm.assureDoor(model.DoorCommand{model.Closed, model.Right})
	} else {
		if nm.chatty {
			log.Printf("I'm somewhere in the middle.\n")
		}
		nm.assureDoor(model.DoorCommand{model.Open, model.Left})
		nm.assureDoor(model.DoorCommand{model.Open, model.Right})
	}
	if nm.chatty {
		log.Println("Current players: ", nm.playersString())
	}
}

func (nm *V23Manager) playersString() (s string) {
	k := nm.findInsertion(nm.myself)
	s = ""
	for i := 0; i < k; i++ {
		s += nm.players[i].p.String() + " "
	}
	if nm.leftDoor == model.Open {
		s += "_"
	} else {
		s += "["
	}
	s += nm.myself.String()
	if nm.rightDoor == model.Open {
		s += "_"
	} else {
		s += "]"
	}
	s += " "
	for i := k; i < len(nm.players); i++ {
		s += nm.players[i].p.String() + " "
	}
	return
}

func (nm *V23Manager) assureDoor(dc model.DoorCommand) {
	switch dc {
	case model.DoorCommand{model.Open, model.Left}:
		if nm.leftDoor == model.Open {
			if nm.chatty {
				log.Printf("Left door already open.\n")
			}
			return
		}
		nm.leftDoor = model.Open
	case model.DoorCommand{model.Open, model.Right}:
		if nm.rightDoor == model.Open {
			if nm.chatty {
				log.Printf("Right door already open.\n")
			}
			return
		}
		nm.rightDoor = model.Open
	case model.DoorCommand{model.Closed, model.Left}:
		if nm.leftDoor == model.Closed {
			if nm.chatty {
				log.Printf("Left door already closed.\n")
			}
			return
		}
		nm.leftDoor = model.Closed
	case model.DoorCommand{model.Closed, model.Right}:
		if nm.rightDoor == model.Closed {
			if nm.chatty {
				log.Printf("Right door already closed.\n")
			}
			return
		}
		nm.rightDoor = model.Closed
	}
	if nm.chDoorCommand == nil {
		log.Panic("The door channel is nil.")
	}
	if nm.chatty {
		log.Printf("Sending door command: %v\n", dc)
	}
	nm.chDoorCommand <- dc
	if nm.chatty {
		log.Printf("Door command %v consumed.\n", dc)
	}
}

func (nm *V23Manager) sayHelloToEveryone() {
	if nm.chatty {
		log.Printf("Me (%v) saying Hello to %d other players.\n",
			nm.Me(), len(nm.players))
	}
	wp := ifc.Player{int32(nm.Me().Id())}
	for _, vp := range nm.players {
		if nm.chatty {
			log.Printf("RPC sending: asking %v to recognize me=%v", vp, nm.Me())
			log.Printf("  nm.ctx %T = %v", nm.ctx, nm.ctx)
			log.Printf("  wp %T = %v", wp, wp)
		}
		if err := vp.c.Recognize(nm.ctx, wp, nm.rpcOpts); err != nil {
			// TODO: Instead of panicing, just drop the player from the players list.
			log.Panic("Recognize failed: ", err)
		}
		if nm.chatty {
			log.Printf("RPC Recognize call completed!")
		}
	}
	if nm.chatty {
		log.Printf("Me (%v) DONE saying Hello.\n", nm.Me())
	}
}

func (nm *V23Manager) sayGoodbyeToEveryone() {
	if nm.chatty {
		log.Println("Saying goodbye to other players.")
	}
	wp := ifc.Player{int32(nm.Me().Id())}
	for _, vp := range nm.players {
		if nm.chatty {
			log.Printf("RPC sending: asking %v to forget me=%v", vp.p, nm.Me())
			log.Printf("  nm.ctx %T = %v", nm.ctx, nm.ctx)
			log.Printf("  wp %T = %v", wp, wp)
		}
		if err := vp.c.Forget(nm.ctx, wp, nm.rpcOpts); err != nil {
			log.Println("Forget failed, but continuing; err=", err)
		}
		if nm.chatty {
			log.Println("Forget call completed.")
		}
	}
}

// Return array of known players.
func (nm *V23Manager) playerNumbers() (list []int) {
	list = []int{}
	rCtx, cancel := context.WithTimeout(nm.ctx, time.Minute)
	defer cancel()
	if nm.chatty {
		log.Printf("Recovering namespace.")
	}
	ns := v23.GetNamespace(rCtx)
	if nm.chatty {
		log.Printf("namespace == %T %v", ns, ns)
	}
	pattern := nm.rootName + "*"
	if nm.chatty {
		log.Printf("Calling glob with %T=%v, pattern=%v\n",
			rCtx, rCtx, pattern)
	}
	c, err := ns.Glob(rCtx, pattern)
	if err != nil {
		log.Printf("ns.Glob(%v) failed: %v", pattern, err)
		return
	}
	if nm.chatty {
		log.Printf("Awaiting response from Glob request.")
	}
	for res := range c {
		if nm.chatty {
			log.Printf("Got a result: %v\n", res)
		}
		switch v := res.(type) {
		case *naming.GlobReplyEntry:
			name := v.Value.Name
			if nm.chatty {
				log.Printf("Raw name is: %v\n", name)
			}
			if name != "" {
				putativeNumber := name[len(nm.rootName):]
				n, err := strconv.ParseInt(putativeNumber, 10, 32)
				if err != nil {
					log.Println(err)
				} else {
					list = append(list, int(n))
				}
				if nm.chatty {
					log.Println("Found player: ", v.Value.Name)
				}
			}
		default:
		}
	}
	if nm.chatty {
		log.Printf("Finished processing glob response.")
	}
	return
}

func (nm *V23Manager) JoinGame(chBc <-chan model.BallCommand) {
	if nm.chatty {
		log.Println("Joining game.")
	}
	nm.chBallCommand = chBc
	for _, id := range nm.initialPlayerNumbers {
		nm.recognizeOther(model.NewPlayer(id))
	}
	if nm.chatty {
		log.Printf("I see %d players.\n", len(nm.players))
	}
	if nm.isGameMaster {
		if chBc != nil {
			log.Panic("game master should not have chBc")
		}
		return
	}
	nm.sayHelloToEveryone()
	nm.checkDoors()
	go nm.run()
	nm.isRunning = true
}

func (nm *V23Manager) run() {
	if nm.chatty {
		log.Println("Starting V23Manager run loop.")
	}
	for {
		select {
		case ch := <-nm.chStop:
			nm.stop()
			ch <- true
			return
		case ch := <-nm.chNoNewBallsOrPeople:
			nm.noNewBallsOrPeople()
			ch <- true
		case bc := <-nm.chBallCommand:
			nm.throwBall(bc)
		case p := <-nm.relay.ChRecognize():
			nm.recognizeOther(p)
		case p := <-nm.relay.ChForget():
			nm.forgetOther(p)
		}
	}
}

func (nm *V23Manager) Quit(id int) {
	for _, vp := range nm.players {
		if vp.p.Id() == id {
			if nm.chatty {
				log.Printf("Killing  %v", vp)
			}
			if err := vp.c.Quit(nm.ctx, nm.rpcOpts); err != nil {
				log.Panic("Quit failed; err=%v", err)
			}
		}
	}
}

func (nm *V23Manager) List() {
	for _, vp := range nm.players {
		log.Printf("%v", vp)
	}
}

func (nm *V23Manager) FireBall(count int) {
	for k := 0; k < count; k++ {
		for _, vp := range nm.players {
			<-time.After(100 * time.Millisecond)
			b := nm.makeBall(vp.p)
			wb := serializeBall(b)
			if nm.chatty {
				log.Printf("Fire ball to %v\n", vp.p)
			}
			if err := vp.c.Accept(nm.ctx, wb, nm.rpcOpts); err != nil {
				log.Panic("Fire ball %v failed; err=%v", b, err)
			}
			if nm.chatty {
				log.Printf("Fire ball %v RPC done.", b)
			}
		}
	}
}

func (nm *V23Manager) makeBall(p *model.Player) *model.Ball {
	dx := rand.Float64()
	dy := rand.Float64()
	sign := rand.Float64()
	if sign >= 0.5 {
		dx = -dx
	}
	mag := math.Sqrt(dx*dx + dy*dy)
	return model.NewBall(p,
		model.Vec{config.MagicX, 0},
		model.Vec{float32(dx / mag), float32(dy / mag)})
}

func (nm *V23Manager) DoMasterCommand(c string) {
	mc := ifc.MasterCommand{Name: c}
	for _, vp := range nm.players {
		if nm.chatty {
			log.Printf("Commanding %v to %v", vp, mc)
		}
		if err := vp.c.DoMasterCommand(nm.ctx, mc, nm.rpcOpts); err != nil {
			log.Panic("Command send failed; err=%v", err)
		}
	}
}

func (nm *V23Manager) SetPauseDuration(pd float32) {
	for _, vp := range nm.players {
		if nm.chatty {
			log.Printf("Setting pause duration to %.2f", pd)
		}
		if err := vp.c.SetPauseDuration(nm.ctx, pd, nm.rpcOpts); err != nil {
			log.Panic("SetPauseDuration failed; err=%v", err)
		}
	}
}

func (nm *V23Manager) SetGravity(g float32) {
	for _, vp := range nm.players {
		if nm.chatty {
			log.Printf("Setting gravity to %.2f", g)
		}
		if err := vp.c.SetGravity(nm.ctx, g, nm.rpcOpts); err != nil {
			log.Panic("SetGravity failed; err=%v", err)
		}
	}
}

// Throw ball either left or right.
func (nm *V23Manager) throwBall(bc model.BallCommand) {
	if nm.chatty {
		log.Printf("v23 manager got ball throw command: %v\n", bc)
	}
	k := nm.findInsertion(nm.myself)
	if bc.D == model.Left {
		// Throw ball left.
		k--
		if k >= 0 {
			nm.sendBallRpc(bc, nm.players[k])
		} else {
			log.Panic("Nobody on left!  Send back to table.")
		}
	} else {
		// Throw ball right.
		if k <= len(nm.players)-1 {
			nm.sendBallRpc(bc, nm.players[k])
		} else {
			log.Panic("Nobody on right!  Send back to table.")
		}
	}
}

func (nm *V23Manager) sendBallRpc(bc model.BallCommand, vp *vPlayer) {
	wb := serializeBall(bc.B)
	if nm.chatty {
		log.Printf("RPC sending: throwing ball %v to %v : %v\n", bc.D, vp.p, vp.c)
	}
	if err := vp.c.Accept(nm.ctx, wb, nm.rpcOpts); err != nil {
		log.Panic("Ball throw %v failed; err=%v", bc.D, err)
	}
	if nm.chatty {
		log.Printf("Ball throw %v RPC done.", bc.D)
	}
}

func serializeBall(b *model.Ball) ifc.Ball {
	wp := ifc.Player{int32(b.Owner().Id())}
	return ifc.Ball{
		wp, b.GetPos().X, b.GetPos().Y, b.GetVel().X, b.GetVel().Y}
}

func (nm *V23Manager) NoNewBallsOrPeople() {
	ch := make(chan bool)
	nm.chNoNewBallsOrPeople <- ch
	<-ch
}

func (nm *V23Manager) noNewBallsOrPeople() {
	if nm.chatty {
		log.Println("********************* No New Balls or people.")
	}
	nm.relay.StopAcceptingData()
	nm.sayGoodbyeToEveryone()
}

func (nm *V23Manager) Stop() {
	ch := make(chan bool)
	nm.chStop <- ch
	<-ch
}

func (nm *V23Manager) stop() {
	if nm.chatty {
		log.Println("v23 calling native shutdown.")
	}
	nm.shutdown()
	if nm.chatty {
		log.Println("v23: closing door command channel.")
	}
	close(nm.chDoorCommand)
	if nm.chatty {
		log.Println("v23 runtime done.")
	}
}
