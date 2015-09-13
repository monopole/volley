package model

import (
	"github.com/monopole/croupier/ifc"
)

type NetManager interface {
	IsRunning() bool
	IsReadyToRun(isGameMaster bool) bool
	ChDoorCommand() <-chan DoorCommand
	ChMasterCommand() <-chan ifc.MasterCommand
	ChKick() <-chan bool
	ChPauseDuration() <-chan float32
	ChGravity() <-chan float32
	ChIncomingBall() <-chan *Ball
	ChQuit() <-chan bool
	Me() *Player
	RunPrep(chBc <-chan BallCommand)
	Run()
	Quit(id int)
	List()
	FireBall(count int)
	DoMasterCommand(c string)
	Kick()
	SetPauseDuration(pd float32)
	SetGravity(g float32)
	NoNewBallsOrPeople()
	Stop()
}
