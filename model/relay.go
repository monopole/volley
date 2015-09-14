package model

import (
	"github.com/monopole/croupier/ifc"
)

type Relay interface {
	ChMasterCommand() <-chan ifc.MasterCommand
	ChPauseDuration() <-chan float32
	ChGravity() <-chan float32
	ChIncomingBall() <-chan *Ball
	ChQuit() <-chan bool
}
