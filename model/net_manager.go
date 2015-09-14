package model

type NetManager interface {
	IsRunning() bool
	GetRelay() Zelay
	GetReady() <-chan bool
	ChDoorCommand() <-chan DoorCommand
	ChKick() <-chan bool
	Me() *Player
	JoinGame(chBc <-chan BallCommand)
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
