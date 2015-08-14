// Relatively dumb V23 service that accepts VOM payloads, converts
// them into game objects, and dumps them asynchronously (so as not to
// block the network thread) onto various send-only channels.
// Dumbness avoids the need v23-dependencies in tests.

package service

import (
	"github.com/monopole/croupier/ifc"
	"github.com/monopole/croupier/model"
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type impl struct {
	chRecognize chan<- *model.Player
	chForget    chan<- *model.Player
	chBall      chan<- *model.Ball
}

func Make(
	chRecognize chan<- *model.Player,
	chForget chan<- *model.Player,
	chBall chan<- *model.Ball,
) ifc.GameBuddyServerMethods {
	return &impl{
		chRecognize: chRecognize,
		chForget:    chForget,
		chBall:      chBall,
	}
}

func (x *impl) Recognize(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	player := model.NewPlayer(int(p.Id))
	go func() {
		x.chRecognize <- player
	}()
	return nil
}

func (x *impl) Forget(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	player := model.NewPlayer(int(p.Id))
	go func() {
		x.chForget <- player
	}()
	return nil
}

func (x *impl) Accept(_ *context.T, _ rpc.ServerCall, b ifc.Ball) error {
	player := model.NewPlayer(int(b.Owner.Id))
	ball := model.NewBall(
		player,
		model.Vec{b.X, b.Y},
		model.Vec{b.Dx, b.Dy})
	go func() {
		x.chBall <- ball
	}()
	return nil
}
