package model

import (
	"github.com/monopole/volley/ifc"
)

type Relay interface {
	ChGravity() <-chan float32
	ChIncomingBall() <-chan *Ball
	ChMasterCommand() <-chan ifc.MasterCommand
	ChPauseDuration() <-chan float32
	ChQuit() <-chan bool
}
