package service

import (
	"github.com/monopole/mutantfortune/ifc"
	"math/rand"
	"sync"
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type impl struct {
	cardOwner int32
	wisdom    []string     // All known fortunes.
	random    *rand.Rand   // To pick a random index in 'wisdom'.
	mu        sync.RWMutex // To safely enable concurrent use.
}

// Makes an implementation.
func Make() ifc.FortuneServerMethods {
	return &impl{
		cardOwner: 0,
		wisdom: []string{
			"You will reach the heights of success.",
			"Conquer your fears or they will conquer you.",
			"Today is your lucky day!",
		},
		random: rand.New(rand.NewSource(99)),
	}
}

func (f *impl) Get(_ *context.T, _ rpc.ServerCall) (blah string, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.wisdom) == 0 {
		return "[empty]", nil
	}
	return f.wisdom[f.random.Intn(len(f.wisdom))], nil
}

func (f *impl) WhoHasCard(_ *context.T, _ rpc.ServerCall) (who int32, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cardOwner, nil
}

func (f *impl) Add(_ *context.T, _ rpc.ServerCall, blah string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.wisdom = append(f.wisdom, blah)
	return nil
}

func (f *impl) SendCardTo(_ *context.T, _ rpc.ServerCall, who int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cardOwner = who
	return nil
}
