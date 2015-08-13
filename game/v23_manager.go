// V23Manager is a peer to other instances of same on the net.
//
// Each device/game/program instance must have a V23Manager.
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
	ctx           *context.T
	shutdown      v23.Shutdown
	rootName      string
	namespaceRoot string
	myNumber      int
	players       []*vPlayer
	chatty        bool

	chInRecognize chan *model.Player
	chInForget    chan *model.Player

	// Anyone holding this channel can tell V23Manager to shut down.
	// Shutdown means close all outgoing channels, stop serving,
	// exit all go routines, then reply on the passed channel.
	chQuit <-chan chan bool
}

func (gm *V23Manager) MyNumber() int {
	return gm.myNumber
}

func (gm *V23Manager) serverName(n int) string {
	return gm.rootName + fmt.Sprintf("%04d", n)
}

func NewV23Manager(
	rootName string, namespaceRoot string,
	chQuit <-chan chan bool) *V23Manager {
	ctx, shutdown := v23.Init()
	if shutdown == nil {
		log.Panic("Why is shutdown nil?")
	}
	//	defer shutdown()
	//		<-signals.ShutdownOnSignals(ctx)
	gm := &V23Manager{
		ctx, shutdown, rootName, namespaceRoot,
		0, []*vPlayer{}, true,
		make(chan *model.Player),
		make(chan *model.Player),
		chQuit}
	return gm
}

func (gm *V23Manager) Initialize(chInBall chan<- *model.Ball) {
	if gm.chatty {
		log.Printf("Scanning namespace %v\n", gm.namespaceRoot)
	}
	v23.GetNamespace(gm.ctx).SetRoots(gm.namespaceRoot)
	numbers := gm.playerNumbers()
	sort.Ints(numbers)
	if len(numbers) == 0 {
		gm.myNumber = 1
	} else {
		gm.myNumber = numbers[len(numbers)-1] + 1
	}
	if gm.chatty {
		log.Printf("My number is %d\n", gm.myNumber)
	}
	gm.registerService(chInBall)
	for _, id := range numbers {
		gm.recognize(model.NewPlayer(id))
	}
	gm.sayHelloToEveryone()
}

func (gm *V23Manager) recognize(p *model.Player) {
	if gm.chatty {
		log.Printf("Recognizing %v\n.", p)
	}
	vp := &vPlayer{p, ifc.GameBuddyClient(gm.serverName(p.Id()))}
	gm.players = append(gm.players, vp)
}

func (gm *V23Manager) sayHelloToEveryone() {
	for _, vp := range gm.players {
		if gm.chatty {
			log.Printf("Saying hello to %v\n", vp.p)
		}
		wp := ifc.Player{int32(vp.p.Id())}
		if err := vp.c.Recognize(gm.ctx, wp, options.SkipServerEndpointAuthorization{}); err != nil {
			log.Panic("Hello failed.")
		}
	}
}

// Modify the game state, and send it to all players, starting with the
// player that's gonna get the card.
func (gm *V23Manager) PassTheCard() {
	if gm.chatty {
		log.Printf("Sending to %v\n", gm.serverName(gm.myNumber)+" "+time.Now().String())
	}
	//	if err := gm.master.SendCardTo(gm.ctx, int32((gm.myNumber+1)%expectedInstances),
	//		options.SkipServerEndpointAuthorization{}); err != nil {
	//		log.Printf("error sending card: %v\n", err)
	//	}
	log.Printf("where i would have sent the card.\n")
}

func (gm *V23Manager) forget(p *model.Player) {
	i := findIndex(len(gm.players),
		func(i int) bool { return gm.players[i].p.Id() == p.Id() })
	if i > -1 {
		if gm.chatty {
			log.Printf("Forgetting %v\n.", p)
		}
		gm.players = append(gm.players[:i], gm.players[i+1:]...)
	} else {
		if gm.chatty {
			log.Printf("Asked to forget %v, but don't know him\n.", p)
		}
	}
}

func (gm *V23Manager) playerCount() int {
	return 0
}

// Return sorted array of known players.
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

// Register a service in the namespace and begin serving.
func (gm *V23Manager) registerService(chInBall chan<- *model.Ball) {
	s := MakeServer(gm.ctx)
	myName := gm.serverName(gm.myNumber)
	if gm.chatty {
		log.Printf("Calling myself %s\n", myName)
	}
	err := s.Serve(
		myName,
		ifc.GameBuddyServer(
			service.Make(gm.chInRecognize, gm.chInForget, chInBall)),
		MakeAuthorizer())
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

func (gm *V23Manager) Run() {
	for {
		select {
		case ch := <-gm.chQuit:
			gm.quit()
			ch <- true
			return
		case p := <-gm.chInRecognize:
			gm.recognize(p)
		case p := <-gm.chInForget:
			gm.forget(p)
		}
	}
}

func (gm *V23Manager) quit() {
	if gm.chatty {
		log.Println("shutting down v23 server...")
	}
	gm.shutdown()

	if gm.chatty {
		log.Println("closing all outgoing channels...")
		log.Println("telling my clients to shutdown...")
		log.Println("all Done...")
	}
}
