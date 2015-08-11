// System and game logic.

// An instance of GameManager is a peer to other instances on the net.
// Each player will have one GameManager.
//
// Each has an embedded V23 service, and is a direct client to the V23
// services held by all the other instances.  It finds all the other
// instances, figures out what it should call itself, and fires off
// go routines to manage data coming in on various channels.
//
// The GameManager is presumably owned by whatever owns the UX event
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

// The number of instances of this program to run in a demo.
// Need an exact count to wire them up properly.
const expectedInstances = 2

func serverName(n int) string {
	return rootName + fmt.Sprintf("%04d", n)
}

type Stringer interface {
	String() string
}

type V23Manager struct {
	ctx      *context.T
	myNumber int // my player number
	master   ifc.GameBuddyClientStub
	chatty   bool    // If true, send fortunes back and forth and log them.  For fun.
	originX  float32 // remember where to put the card
	originY  float32
}

func (gm *V23Manager) MyNumber() int {
	return gm.myNumber
}

func NewV23Manager(ctx *context.T) *V23Manager {
	ctx, shutdown := v23.Init()
	if shutdown == nil {
		log.Panic("Why is shutdown nil?")
	}
	//	defer shutdown()
	//		<-signals.ShutdownOnSignals(ctx)
	gm := &V23Manager{ctx, 0, nil, true, 0, 0}
	gm.initialize()
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
