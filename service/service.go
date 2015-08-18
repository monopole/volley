// Relatively dumb V23 service that accepts VOM payloads, converts
// them into game objects, and dumps them asynchronously (so as not to
// block the network thread) onto various receive-only channels.
// Dumbness avoids the need v23-dependencies in tests.

package service

import (
	"github.com/monopole/croupier/config"
	"github.com/monopole/croupier/ifc"
	"github.com/monopole/croupier/model"
	"log"
	"sync"
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type Relay struct {
	chRecognize    chan *model.Player
	chForget       chan *model.Player
	chIncomingBall chan *model.Ball
	acceptingData  bool
	mu             sync.RWMutex
}

func MakeRelay() *Relay {
	r := &Relay{}
	r.chRecognize = make(chan *model.Player)
	r.chForget = make(chan *model.Player)
	r.chIncomingBall = make(chan *model.Ball)
	r.acceptingData = true
	return r
}

// Don't call this until after someone calls v23.shutdown, or incoming
// RPC's will attempt to write to closed channels.  Calling this will
// presumably unblock anything waiting for, say, a new Ball.
//
// With some extra work and a mutex member, we could add a more
// complex lifecycle to turn off the relay and turn it back on at
// will, to support leaving and re-entering the game without losing
// one's number, name, and place in the mount table.  If off mode,
// data from incoming RPC's would just get dropped on the floor
// instead of placed on the channel.
func (x *Relay) Close() {
	x.mu.Lock()
	defer x.mu.Unlock()
	close(x.chRecognize)
	close(x.chForget)
	close(x.chIncomingBall)
	x.acceptingData = false
	if config.Chatty {
		log.Printf("Refusing all data.")
	}
}

func (x *Relay) ChRecognize() <-chan *model.Player {
	return x.chRecognize
}

func (x *Relay) ChForget() <-chan *model.Player {
	return x.chForget
}

func (x *Relay) ChIncomingBall() <-chan *model.Ball {
	return x.chIncomingBall
}

func (x *Relay) Recognize(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	go func() {
		x.mu.Lock()
		defer x.mu.Unlock()
		if x.acceptingData {
			if config.Chatty {
				log.Printf("RPC received: accepting recognize request from player %v", p)
			}
			player := model.NewPlayer(int(p.Id))
			if config.Chatty {
				log.Printf("Enchanneling newly recognized player = %v", player)
			}
			x.chRecognize <- player
		} else {
			log.Printf("Discarding recognize request from player %v", p)
		}
	}()
	return nil
}

func (x *Relay) Forget(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	go func() {
		x.mu.Lock()
		defer x.mu.Unlock()
		if x.acceptingData {
			if config.Chatty {
				log.Printf("RPC received: accepting forget request from player %v", p)
			}
			player := model.NewPlayer(int(p.Id))
			if config.Chatty {
				log.Printf("Enchanneling player to forget = %v", player)
			}
			x.chForget <- player
		} else {
			log.Printf("Discarding forget request from player %v", p)
		}
	}()
	return nil
}

func (x *Relay) Accept(_ *context.T, _ rpc.ServerCall, b ifc.Ball) error {
	go func() {
		x.mu.Lock()
		defer x.mu.Unlock()
		if x.acceptingData {
			if config.Chatty {
				log.Printf("RPC received: accepting ball %v", b)
			}
			player := model.NewPlayer(int(b.Owner.Id))
			ball := model.NewBall(
				player,
				model.Vec{b.X, b.Y},
				model.Vec{b.Dx, b.Dy})
			if config.Chatty {
				log.Printf("Enchanneling ball = %v", ball)
			}
			x.chIncomingBall <- ball
		} else {
			log.Printf("Discarding ball.")
		}
	}()
	return nil
}
