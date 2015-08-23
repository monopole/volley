package model

type ExecCommand int

const (
	ExecError ExecCommand = iota
	ExecStart
	ExecRandomImpulse
	ExecPaint
)

type DoorState int

const (
	Open DoorState = iota
	Closed
)

func (s DoorState) String() string {
	if s == Open {
		return "open"
	} else {
		return "closed"
	}
}

type Direction int

const (
	Left Direction = iota
	Right
)

func (s Direction) String() string {
	if s == Left {
		return "left"
	} else {
		return "right"
	}
}

type DoorCommand struct {
	S DoorState
	D Direction
}

func (dc DoorCommand) String() string {
	return dc.S.String() + "-" + dc.D.String()
}
