package screen

// See
// https://github.com/golang/mobile/blob/master/example/basic/main.go

import (
	"github.com/monopole/croupier/model"
	"github.com/monopole/croupier/sprite"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
)

type Screen struct {
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	zball    *sprite.Ball
	opaque   float32
	green    float32
	red      float32
	blue     float32
	gray     float32
	width    float32
	height   float32
}

func NewScreen() *Screen {
	return &Screen{}
}

func (s *Screen) Start() {
	var err error
	s.opaque = 1.0
	s.red = 0.1
	s.green = 0.1
	s.blue = 0.1

	s.program, err = glutil.CreateProgram(sprite.VertexShader, sprite.FragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	s.zball = sprite.NewBall()
	s.position = gl.GetAttribLocation(s.program, "position")
	s.color = gl.GetUniformLocation(s.program, "color")
	s.offset = gl.GetUniformLocation(s.program, "offset")
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
	gl.ClearColor(s.red, s.green, s.blue, s.opaque)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(s.program)

	s.gray += 0.01
	if s.gray > 1 {
		s.gray = 0
	}
	gl.Uniform4f(s.color, s.gray, 0, s.gray, s.opaque)

	if len(balls) > 0 {
		b := balls[0]
		gl.Uniform2f(s.offset, b.GetPos().X/s.width, b.GetPos().Y/s.height)
	}
	gl.EnableVertexAttribArray(s.position)
	gl.VertexAttribPointer(s.position, sprite.CoordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, sprite.VertexCount)

	gl.DisableVertexAttribArray(s.position)

	// debug.DrawFPS(c)
}

func (s *Screen) Stop() {
	gl.DeleteProgram(s.program)
	s.zball.DeleteBuffer()
}
