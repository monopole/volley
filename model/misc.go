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

type Direction int

const (
	Left Direction = iota
	Right
)

type DoorCommand struct {
	S DoorState
	D Direction
}

func (dc DoorCommand) String() string {
	if dc.S == Open {
		if dc.D == Left {
			return "open-left"
		} else {
			return "open-right"
		}
	} else {
		if dc.D == Left {
			return "close-left"
		} else {
			return "close-right"
		}
	}
}
