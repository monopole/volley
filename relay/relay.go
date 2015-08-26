// V23 service that accepts VOM payloads, converts them into game
// objects, and dumps them asynchronously (so as not to block the
// network thread) on receive-only channels.

package relay

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
	chRecognize     chan *model.Player
	chForget        chan *model.Player
	chBall          chan *model.Ball
	chQuit          chan bool
	chKick          chan bool
	chMasterCommand chan ifc.MasterCommand
	chPauseDuration chan float32
	chGravity       chan float32
	acceptingData   bool
	mu              sync.RWMutex
}

func MakeRelay() *Relay {
	r := &Relay{}
	r.chRecognize = make(chan *model.Player)
	r.chForget = make(chan *model.Player)
	r.chBall = make(chan *model.Ball)
	r.chQuit = make(chan bool)
	r.chKick = make(chan bool)
	r.chMasterCommand = make(chan ifc.MasterCommand)
	r.chPauseDuration = make(chan float32)
	r.chGravity = make(chan float32)
	r.acceptingData = true
	if config.Chatty {
		log.Printf("Made Relay.")
	}
	return r
}

// This is currently undoable.  Would need logic elsewhere
// to be able to turn it back on again.
func (r *Relay) StopAcceptingData() {
	if config.Chatty {
		log.Printf("Relay: no more data...")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.acceptingData = false
	// Closing the channel sends a nil, which unblocks, but the receiver
	// must deal with nil.  Setting the channel nil means nothing will
	// select it (it blocks).
	r.chRecognize = nil
	r.chForget = nil
	r.chBall = nil
	if config.Chatty {
		log.Printf("Relay: no more data!")
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

func (r *Relay) ChQuit() <-chan bool {
	return r.chQuit
}

func (r *Relay) ChMasterCommand() <-chan ifc.MasterCommand {
	return r.chMasterCommand
}

func (r *Relay) ChKick() <-chan bool {
	return r.chKick
}

func (r *Relay) ChPauseDuration() <-chan float32 {
	return r.chPauseDuration
}

func (r *Relay) ChGravity() <-chan float32 {
	return r.chGravity
}

func (r *Relay) DoMasterCommand(_ *context.T, _ rpc.ServerCall, mc ifc.MasterCommand) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingData {
			if config.Chatty {
				log.Printf("Relay: MasterCommand = %v", mc)
			}
			r.chMasterCommand <- mc
			if config.Chatty {
				log.Printf("Relay: Passed in mc.")
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: Discarding mc.")
			}
		}
	}()
	return nil
}

func (r *Relay) SetPauseDuration(_ *context.T, _ rpc.ServerCall, p float32) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingData {
			if config.Chatty {
				log.Printf("Relay: Pause duration = %.2f", p)
			}
			r.chPauseDuration <- p
			if config.Chatty {
				log.Printf("Relay: Passed in pause.")
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: Discarding pause.")
			}
		}
	}()
	return nil
}

func (r *Relay) SetGravity(_ *context.T, _ rpc.ServerCall, g float32) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingData {
			if config.Chatty {
				log.Printf("Relay: gravity = %.2f", g)
			}
			r.chGravity <- g
			if config.Chatty {
				log.Printf("Relay: Passed in gravity.")
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: Discarding gravity.")
			}
		}
	}()
	return nil
}

func (r *Relay) Kick(_ *context.T, _ rpc.ServerCall) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingData {
			if config.Chatty {
				log.Printf("Relay: Got a kick!.")
			}
			r.chKick <- true
			if config.Chatty {
				log.Printf("Relay: Passed kick to ch.")
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: Discarding kick request.")
			}
		}
	}()
	return nil
}

func (r *Relay) Quit(_ *context.T, _ rpc.ServerCall) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingData {
			if config.Chatty {
				log.Printf("Relay: Got a quit!.")
			}
			r.chQuit <- true
			if config.Chatty {
				log.Printf("Relay: Passed quit to ch.")
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: Discarding quit request.")
			}
		}
	}()
	return nil
}

func (r *Relay) Recognize(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingData {
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
		if r.acceptingData {
			player := model.NewPlayer(int(p.Id))
			if config.Chatty {
				log.Printf("Relay: Must forget player %v", player)
			}
			r.chForget <- player
			if config.Chatty {
				log.Printf("Relay: Forget %v consumed.", player)
			}
		} else {
			if config.Chatty {
				log.Printf("Relay: Discarding forget request from player %v", p)
			}
		}
	}()
	return nil
}

func (r *Relay) Accept(_ *context.T, _ rpc.ServerCall, b ifc.Ball) error {
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.acceptingData {
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
