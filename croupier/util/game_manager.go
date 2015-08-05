package util

import (
	"github.com/monopole/mutantfortune/ifc"
	"github.com/monopole/mutantfortune/server/util"
	"github.com/monopole/mutantfortune/service"
	"log"
	"strconv"
	"time"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/options"
	_ "v.io/x/ref/runtime/factories/generic"
)

const rootName = "croupier"

// The number of instances of this program to run in a demo.
// Need an exact count to wire them up properly.
const expectedInstances = 2

func serverName(n int) string {
	return rootName + strconv.Itoa(n)
}

type GameManager struct {
	ctx      *context.T
	myNumber int // my player number
	master   ifc.FortuneClientStub
	originX  float32 // remember where to put the card
	originY  float32
	chatty   bool // If true, send fortunes back and forth and log them.  For fun.
}

func (gm *GameManager) MyNumber() int {
	return gm.myNumber
}

func (gm *GameManager) GetOriginX() float32 {
	return gm.originX
}

func (gm *GameManager) GetOriginY() float32 {
	return gm.originY
}

func (gm *GameManager) SetOrigin(x, y float32) {
	gm.originX = x
	gm.originY = y
}

func NewGameManager(ctx *context.T) *GameManager {
	gm := &GameManager{ctx, 0, nil, 0, 0, false}
	gm.initialize()
	return gm
}

func (gm *GameManager) initialize() {
	v23.GetNamespace(gm.ctx).SetRoots("/104.197.96.113:3389")

	// If there are no croupiers, I register as server0.
	// If there is one croupier already, I register as server1, etc.
	// Only the first server to register will actuall be used as a server.
	// The remaining instances  act as overblown presence counters.
	gm.myNumber = gm.croupierCount()

	gm.registerService()

	// No matter who I am, I am a client to server0.
	gm.master = ifc.FortuneClient(serverName(0))
}

// Scan mounttable for count of services matching "croupier*"
func (gm *GameManager) croupierCount() (count int) {
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

// Register a service in the namespace.
func (gm *GameManager) registerService() {
	s := util.MakeServer(gm.ctx)
	myName := serverName(gm.myNumber)
	log.Printf("Calling myself %s\n", myName)
	err := s.Serve(myName, ifc.FortuneServer(service.Make()), util.MakeAuthorizer())
	if err != nil {
		log.Panic("Error serving service: ", err)
	}
}

func (gm *GameManager) WhoHasTheCard() int {
	if gm.chatty {
		fortune, err := gm.master.Get(gm.ctx, options.SkipServerEndpointAuthorization{})
		if err != nil {
			log.Printf("error getting fortune: %v", err)
			return 0
		}
		log.Println(fortune)
	}
	who, _ := gm.master.WhoHasCard(gm.ctx, options.SkipServerEndpointAuthorization{})
	return int(who)
}

func (gm *GameManager) PassTheCard() {
	if gm.chatty {
		if err := gm.master.Add(gm.ctx, serverName(gm.myNumber)+" "+time.Now().String(),
			options.SkipServerEndpointAuthorization{}); err != nil {
			log.Printf("error adding fortune: %v\n", err)
		}
	}
	if err := gm.master.SendCardTo(gm.ctx, int32((gm.myNumber+1)%expectedInstances),
		options.SkipServerEndpointAuthorization{}); err != nil {
		log.Printf("error sending card: %v\n", err)
	}
}
