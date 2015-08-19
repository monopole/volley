package screen

// See
// https://github.com/golang/mobile/blob/master/example/basic/main.go

import (
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/sprite"
	"golang.org/x/mobile/gl"
)

const (
	opaque = 1.0
	red    = 0.1
	green  = 0.1
	blue   = 0.1
)

type Screen struct {
	zball  []*sprite.Ball
	gray   float32
	width  float32
	height float32
}

func NewScreen() *Screen {
	return &Screen{}
}

func (s *Screen) Start() {

	s.zball = []*sprite.Ball{sprite.NewBall()}
	//	s.zball = []*sprite.Ball{sprite.NewBall(), sprite.NewBall()}
}

func (s *Screen) ReSize(width float32, height float32) {
	s.width = width
	s.height = height
}

func (s *Screen) Width() float32 {
	return s.width
}

func (s *Screen) Height() float32 {
	return s.height
}

func (s *Screen) Paint(balls []*model.Ball) {
	gl.ClearColor(red, green, blue, opaque)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	s.gray += 0.01
	if s.gray > 1 {
		s.gray = 0
	}

	if len(balls) > 0 {
		b := balls[0]
		x := b.GetPos().X / s.width
		y := b.GetPos().Y / s.height
		// Experimenting with multiple balls not working.
		s.zball[0].Draw(s.gray, x, y)
		//		s.zball[1].Draw(s.gray, x+100, y+10)
	}

	// debug.DrawFPS(c)
}

func (s *Screen) Stop() {
	s.zball[0].Delete()
	//	s.zball[1].Delete()
}
