// System and game logic.

// An instance of V23Manager is a peer to other instances on the net.
// Each player will have one V23Manager.
//
// Each has an embedded V23 service, and is a direct client to the V23
// services held by all the other instances.  It finds all the other
// instances, figures out what it should call itself, and fires off
// go routines to manage data coming in on various channels.
//
// The V23Manager is presumably owned by whatever owns the UX event
// loop,
//
// During play, UX or underlying android/iOS events may trigger calls
// to other V23 services, Likewise, an incoming RPC may change data
// held by the manager, to ultimately impact the UX (e.g. a card is
// passed in by another player).

package game

import (
	"fmt"
	"github.com/monopole/croupier/ifc"
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/service"
	"log"
	"time"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	//	"v.io/v23/options"
	_ "v.io/x/ref/runtime/factories/generic"
)

const rootName = "croupier/player"
const namespaceRoot = "/104.197.96.113:3389"

func serverName(n int) string {
	return rootName + fmt.Sprintf("%04d", n)
}

type V23Manager struct {
	ctx      *context.T
	myNumber int
	players  []*model.Player
	master   ifc.GameBuddyClientStub
	chatty   bool

	// V23Manager sends balls out on this channel.
	chBall chan model.Ball

	//chForget    chan model.Player
	//chRecognize chan model.Player

	// Anyone holding this channel can tell V23Manager to shut down.
	// Shutdown means close all outgoing channels, stop serving, and
	// exit all go routines.
	chQuit <-chan chan bool
}

func (gm *V23Manager) MyNumber() int {
	return gm.myNumber
}

func NewV23Manager(ctx *context.T, chQuit <-chan chan bool) *V23Manager {
	ctx, shutdown := v23.Init()
	if shutdown == nil {
		log.Panic("Why is shutdown nil?")
	}
	//	defer shutdown()
	//		<-signals.ShutdownOnSignals(ctx)
	gm := &V23Manager{
		ctx, 0, nil, true,
		make(chan model.Ball),
		// make(chan model.Player),
		// make(chan model.Player),
		chQuit}
	//gm.initialize()
	return gm
}

func (gm *V23Manager) initialize() {
	v23.GetNamespace(gm.ctx).SetRoots(namespaceRoot)

	gm.myNumber = gm.playerCount()

	// If there are no players, I register as player 1.  If there is one
	// player already, I register as player 2, etc.
	gm.registerService()

	// No matter who I am, I am a client to server0.
	gm.master = ifc.GameBuddyClient(serverName(0))
}

// Scan mounttable for count of services matching "{rootName}*"
func (gm *V23Manager) playerCount() (count int) {
	count = 0
	rCtx, cancel := context.WithTimeout(gm.ctx, time.Minute)
	defer cancel()
	ns := v23.GetNamespace(rCtx)
	pattern := rootName + "*"
	c, err := ns.Glob(rCtx, pattern)
	if err != nil {
		log.Printf("ns.Glob(%q) failed: %v", pattern, err)
		return
	}
	for res := range c {
		switch v := res.(type) {
		case *naming.GlobReplyEntry:
			if v.Value.Name != "" {
				count++
				if gm.chatty {
					log.Println(v.Value.Name)
				}
			}
		default:
		}
	}
	return
}

func (gm *V23Manager) run() {
	for {
		select {
		case ch := <-gm.chQuit:
			gm.shutdown()
			ch <- true
			return
		}
	}
}

func (gm *V23Manager) shutdown() {
	log.Println("closing all outgoing channels...")
	log.Println("telling my clients to shutdown...")
	log.Println("shutting down server...")
	log.Println("all Done...")
}

// Register a service in the namespace and begin serving.
func (gm *V23Manager) registerService() {
	s := MakeServer(gm.ctx)
	myName := serverName(gm.myNumber)
	log.Printf("Calling myself %s\n", myName)
	err := s.Serve(myName, ifc.GameBuddyServer(service.Make()), MakeAuthorizer())
	if err != nil {
		log.Panic("Error serving service: ", err)
	}
}

func findIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func (gm *V23Manager) removePlayer(p *model.Player) {
	i := findIndex(len(gm.players), func(i int) bool { return gm.players[i].Id == p.Id })
	if i > -1 {
		gm.players = append(gm.players[:i], gm.players[i+1:]...)
	}
}

func (gm *V23Manager) WhoHasTheCard() int {
	// who, _ := gm.master.WhoHasCard(gm.ctx, options.SkipServerEndpointAuthorization{})
	who := 1
	return int(who)
}

// Modify the game state, and send it to all players, starting with the
// player that's gonna get the card.
func (gm *V23Manager) PassTheCard() {
	if gm.chatty {
		log.Printf("Sending to %v\n", serverName(gm.myNumber)+" "+time.Now().String())
	}
	//	if err := gm.master.SendCardTo(gm.ctx, int32((gm.myNumber+1)%expectedInstances),
	//		options.SkipServerEndpointAuthorization{}); err != nil {
	//		log.Printf("error sending card: %v\n", err)
	//	}
	log.Printf("where i would have sent the card.\n")
}
