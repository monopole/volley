package service

import (
	"github.com/monopole/croupier/ifc"
	"math/rand"
	"sync"
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type impl struct {
	players []int32
	state   ifc.GameState
	mu      sync.RWMutex
}

func Make() ifc.GameBuddyServerMethods {
	return &impl{
		players: []int32{},
		state:   {0, 0},
	}
}

func (f *impl) Recognize(_ *context.T, _ rpc.ServerCall, n int32) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	f.players = append(f.players, n)
	return nil
}

func (f *impl) Forget(_ *context.T, _ rpc.ServerCall, n int32) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	f.removePlayer(n)
	return nil
}

func (f *impl) removePlayer(n int32) {
	// Implement.
}

func (f *impl) PutGameState(_ *context.T, _ rpc.ServerCall, incoming GameState) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	f.state = incoming
	return nil
}

func (f *impl) GetGameState(_ *context.T, _ rpc.ServerCall) (s GameState, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state, nil
}
