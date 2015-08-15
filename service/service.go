// Relatively dumb V23 service that accepts VOM payloads, converts
// them into game objects, and dumps them asynchronously (so as not to
// block the network thread) onto various receive-only channels.
// Dumbness avoids the need v23-dependencies in tests.

package service

import (
	"github.com/monopole/croupier/ifc"
	"github.com/monopole/croupier/model"
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type Relay struct {
	chRecognize    chan *model.Player
	chForget       chan *model.Player
	chIncomingBall chan *model.Ball
}

func MakeRelay() *Relay {
	return &Relay{
		make(chan *model.Player),
		make(chan *model.Player),
		make(chan *model.Ball)}
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
	close(x.chRecognize)
	close(x.chForget)
	close(x.chIncomingBall)
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
	player := model.NewPlayer(int(p.Id))
	go func() {
		x.chRecognize <- player
	}()
	return nil
}

func (x *Relay) Forget(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	player := model.NewPlayer(int(p.Id))
	go func() {
		x.chForget <- player
	}()
	return nil
}

func (x *Relay) Accept(_ *context.T, _ rpc.ServerCall, b ifc.Ball) error {
	player := model.NewPlayer(int(b.Owner.Id))
	ball := model.NewBall(
		player,
		model.Vec{b.X, b.Y},
		model.Vec{b.Dx, b.Dy})
	go func() {
		x.chIncomingBall <- ball
	}()
	return nil
}
