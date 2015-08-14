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
	ctx           *context.T
	shutdown      v23.Shutdown
	rootName      string
	namespaceRoot string
	relay         *service.Relay
	chatty        bool
	me            *model.Player
	players       []*vPlayer
	chQuit        chan chan bool
}

func NewV23Manager(
	rootName string, namespaceRoot string) *V23Manager {
	ctx, shutdown := v23.Init()
	if shutdown == nil {
		log.Panic("shutdown nil")
	}
	return &V23Manager{
		ctx, shutdown,
		rootName, namespaceRoot,
		nil, true, nil, []*vPlayer{},
		make(chan chan bool),
	}
}

func (gm *V23Manager) Initialize() {
	if gm.chatty {
		log.Printf("Scanning namespace %v\n", gm.namespaceRoot)
	}
	v23.GetNamespace(gm.ctx).SetRoots(gm.namespaceRoot)

	numbers := gm.playerNumbers()
	sort.Ints(numbers)
	myId := 1
	if len(numbers) > 0 {
		myId = numbers[len(numbers)-1] + 1
	}
	gm.me = model.NewPlayer(myId)
	if gm.chatty {
		log.Printf("I am %v\n", gm.me)
	}

	s := MakeServer(gm.ctx)
	myName := gm.serverName(gm.Me().Id())
	if gm.chatty {
		log.Printf("Calling myself %s\n", myName)
	}
	gm.relay = service.MakeRelay()
	err := s.Serve(
		myName,
		ifc.GameBuddyServer(gm.relay),
		MakeAuthorizer())
	if err != nil {
		log.Panic("Error serving service: ", err)
	}

	for _, id := range numbers {
		gm.recognize(model.NewPlayer(id))
	}
	gm.sayHelloToEveryone()
}

func (gm *V23Manager) Quitter() chan<- chan bool {
	return gm.chQuit
}

func (gm *V23Manager) chIncomingBall() <-chan *model.Ball {
	return gm.relay.ChBall()
}

func (gm *V23Manager) Me() *model.Player {
	return gm.me
}

func (gm *V23Manager) serverName(n int) string {
	return gm.rootName + fmt.Sprintf("%04d", n)
}

func (gm *V23Manager) recognize(p *model.Player) {
	if gm.chatty {
		log.Printf("%v recognizing %v.", gm.Me(), p)
	}
	vp := &vPlayer{p, ifc.GameBuddyClient(gm.serverName(p.Id()))}
	gm.players = append(gm.players, vp)
}

func (gm *V23Manager) sayHelloToEveryone() {
	wp := ifc.Player{int32(gm.Me().Id())}
	for _, vp := range gm.players {
		if gm.chatty {
			log.Printf("Asking %v to recognize %v\n", vp.p, gm.Me())
		}
		if err := vp.c.Recognize(
			gm.ctx, wp, options.SkipServerEndpointAuthorization{}); err != nil {
			log.Panic("Recognize failed.")
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
			log.Printf("Asking %v to forget %v\n", vp.p, gm.Me())
		}
		if err := vp.c.Forget(
			gm.ctx, wp, options.SkipServerEndpointAuthorization{}); err != nil {
			log.Panic("Forget failed.")
		}
	}
}

func (gm *V23Manager) forget(p *model.Player) {
	i := findIndex(len(gm.players),
		func(i int) bool { return gm.players[i].p.Id() == p.Id() })
	if i > -1 {
		if gm.chatty {
			log.Printf("%v forgetting %v.\n", gm.Me(), p)
		}
		gm.players = append(gm.players[:i], gm.players[i+1:]...)
	} else {
		if gm.chatty {
			log.Printf("Asked to forget %v, but don't know him\n.", p)
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

func findIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func (gm *V23Manager) Run() {
	if gm.chatty {
		log.Println("V23 Running.")
	}
	for {
		select {
		case ch := <-gm.chQuit:
			gm.quit()
			ch <- true
			return
		case p := <-gm.relay.ChRecognize():
			gm.recognize(p)
		case p := <-gm.relay.ChForget():
			gm.forget(p)
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
}
