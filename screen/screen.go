package screen

// Based on https://github.com/golang/mobile/blob/master/example/basic/main.go

import (
	"encoding/binary"
	"github.com/monopole/croupier/model"
	// "golang.org/x/mobile/event/size"
	//	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
)

type Screen struct {
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	buf      gl.Buffer
	green    float32
	red      float32
	blue     float32
	gray     float32
	width    float32
	height   float32
}

func NewScreen() *Screen {
	//	return &Screen{0, 0, nil, nil, nil, nil, nil, 0, 0, 0, 0}
	return &Screen{}
}

func (s *Screen) Start() {
	var err error

	s.red = 0.1
	s.green = 0.8
	s.blue = 0.4

	s.program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	s.buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, s.buf)
	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	s.position = gl.GetAttribLocation(s.program, "position")
	s.color = gl.GetUniformLocation(s.program, "color")
	s.offset = gl.GetUniformLocation(s.program, "offset")

	// TODO(crawshaw): the debug package needs to put GL state init here
	// Can this be an app.RegisterFilter call now??
}

func (s *Screen) ReSize(width float32, height float32) {
	s.width = width   // touchX/float32(sz.WidthPx),
	s.height = height // where sz is size.Event
}

func (s *Screen) Paint(balls []*model.Ball) {

	gl.ClearColor(s.red, s.green, s.blue, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(s.program)

	s.gray += 0.01
	if s.gray > 1 {
		s.gray = 0
	}
	// Color the triangle
	gl.Uniform4f(s.color, s.gray, 0, s.gray, 1)

	// Move the triangle
	b := balls[0]
	gl.Uniform2f(s.offset, b.GetPos().X/s.width, b.GetPos().Y/s.height)

	gl.BindBuffer(gl.ARRAY_BUFFER, s.buf)
	gl.EnableVertexAttribArray(s.position)
	gl.VertexAttribPointer(s.position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)

	gl.DisableVertexAttribArray(s.position)

	// debug.DrawFPS(c)
}

func (s *Screen) Stop() {
	gl.DeleteProgram(s.program)
	gl.DeleteBuffer(s.buf)
}

var triangleData = f32.Bytes(binary.LittleEndian,
	// x, y, z, in percentage distance from origin to window border
	-0.4, 0.2, 0.0,
	0.4, 0.2, 0.0,
	0.4, -0.2, 0.0,
	-0.4, -0.2, 0.0,
)

const (
	coordsPerVertex = 3
	vertexCount     = 4
)

const vertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
