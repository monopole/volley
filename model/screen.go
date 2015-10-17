package model

type Screen interface {
	SetDrawContext(interface{}) error
	Start()
	ReSize(width float32, height float32)
	Width() float32
	Height() float32
	Clear()
	Paint(balls []*Ball)
	Stop()
}
