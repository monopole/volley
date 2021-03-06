package model

type NetManager interface {
	IsRunning() bool
	GetRelay() Relay
	GetReady() <-chan bool
	ChDoorCommand() <-chan DoorCommand
	Me() *Player
	JoinGame(chBc <-chan BallCommand)
	Quit(id int)
	List()
	FireBall(count int)
	DoMasterCommand(c string)
	SetPauseDuration(pd float32)
	SetGravity(g float32)
	NoNewBallsOrPeople()
	Stop()
}
