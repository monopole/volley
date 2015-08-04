package service

import (
	"github.com/monopole/mutantfortune/ifc"
	"math/rand"
	"sync"
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type impl struct {
	wisdom []string     // All known fortunes.
	random *rand.Rand   // To pick a random index in 'wisdom'.
	mu     sync.RWMutex // To safely enable concurrent use.
}

// Makes an implementation.
func Make() ifc.FortuneServerMethods {
	return &impl{
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

func (f *impl) Add(_ *context.T, _ rpc.ServerCall, blah string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.wisdom = append(f.wisdom, blah)
	return nil
}
