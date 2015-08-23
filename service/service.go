// Relatively dumb V23 service that accepts VOM payloads, converts
// them into game objects, and dumps them asynchronously (so as not to
// block the network thread) on receive-only channels.
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
	chRecognize      chan *model.Player
	chForget         chan *model.Player
	chBall           chan *model.Ball
	acceptingBalls   bool
	acceptingPlayers bool
	mu               sync.RWMutex
}

func MakeRelay() *Relay {
	r := &Relay{}
	r.chRecognize = make(chan *model.Player)
	r.chForget = make(chan *model.Player)
	r.chBall = make(chan *model.Ball)
	r.acceptingBalls = true
	r.acceptingPlayers = true
	if config.Chatty {
		log.Printf("Made Relay.")
	}
	return r
}

func (r *Relay) StopAcceptingPlayers() {
	if config.Chatty {
		log.Printf("Relay: no more players...")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.acceptingPlayers = false
	// Closing the channel sends a nil, which unblocks, but the receiver
	// must deal with nil.
	// Setting the channel nil means nothing will select it (it blocks).
	r.chRecognize = nil
	if config.Chatty {
		log.Printf("Relay: no more players!")
	}
}

func (r *Relay) StopAcceptingBalls() {
	if config.Chatty {
		log.Printf("Relay: no more balls...")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.acceptingBalls = false
	r.chBall = nil
	if config.Chatty {
		log.Printf("Relay: no more balls!")
	}
}

// Don't call this until after someone calls v23.shutdown, or incoming
// RPC's will attempt to write to closed channels.  Calling this will
// presumably unblock anything waiting for, say, a new Ball.
//
// With some extra work and a mutex member, we could add a more
// complex lifecycle to turn off the relay and turn it back on at
// will, to support leaving and re-entering the game without losing
// one's number, name, and place in the mount table.  In "idle" mode,
// data from incoming RPC's would be dropped on the floor
// instead of placed on the channel.
func (r *Relay) Stop() {
	if config.Chatty {
		log.Printf("Relay: Grabbing lock.")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.acceptingBalls {
		log.Panic("Stop attempt while still accepting balls!")
	}
	if r.acceptingPlayers {
		log.Panic("Stop attempt while still accepting players!")
	}
	close(r.chForget)
	if config.Chatty {
		log.Printf("Relay: closed.")
	}
}

func (r *Relay) ChRecognize() <-chan *model.Player {
	return r.chRecognize
}

func (r *Relay) ChForget() <-chan *model.Player {
	return r.chForget
}

func (r *Relay) ChIncomingBall() <-chan *model.Ball {
	return r.chBall
}

func (r *Relay) Recognize(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingPlayers {
			player := model.NewPlayer(int(p.Id))
			if config.Chatty {
				log.Printf("Relay: Must recognize player %v", player)
			}
			r.chRecognize <- player
			if config.Chatty {
				log.Printf("Relay: Recognize %v consumed.", player)
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: Discarding recognize request from player %v", p)
			}
		}
	}()
	return nil
}

func (r *Relay) Forget(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		player := model.NewPlayer(int(p.Id))
		if config.Chatty {
			log.Printf("Relay: Must forget player %v", player)
		}
		r.chForget <- player
		if config.Chatty {
			log.Printf("Relay: Forget %v consumed.", player)
		}
	}()
	return nil
}

func (r *Relay) Accept(_ *context.T, _ rpc.ServerCall, b ifc.Ball) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingBalls {
			player := model.NewPlayer(int(b.Owner.Id))
			ball := model.NewBall(
				player,
				model.Vec{b.X, b.Y},
				model.Vec{b.Dx, b.Dy})
			if config.Chatty {
				log.Printf("Relay: accepting ball %v", ball)
			}
			r.chBall <- ball
			if config.Chatty {
				log.Printf("Relay: accepted  %v", ball)
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: dropping ball on floor.")
			}
		}
	}()
	return nil
}
