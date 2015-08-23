package screen

// See
// https://github.com/golang/mobile/blob/master/example/basic/main.go

import (
	"encoding/binary"
	"github.com/monopole/croupier/model"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
	"math"
)

const (
	extraBalls      = 10
	coordsPerVertex = 3
	vertexCount     = 4
	opaque          = 1 // 0 is transparent
	bgRed           = 0.1
	bgGreen         = 0.1
	bgBlue          = 0.1

	// See coords.txt
	vertexShader = `#version 100
uniform vec2 jrOffset;
attribute vec4 jrPosition;
void main() {
	vec4 offset4 = vec4(2.0*jrOffset.x-1.0, 1.0-2.0*jrOffset.y, 0, 0);
	gl_Position = jrPosition + offset4;
}`

	fragmentShader = `#version 100
precision mediump float;
uniform vec4 jrColor;
void main() {
	gl_FragColor = jrColor;
}`
)

type Color struct {
	R, G, B float32
}

var playerColors = []Color{
	Color{255, 255, 255}, // white
	Color{0, 87, 231},    // google blue
	Color{214, 45, 32},   // google red
	Color{255, 167, 0},   // google orange
	Color{0, 135, 68},    // google green
	Color{255, 0, 255},   // magenta
	Color{0, 255, 255},   // cyan
	Color{218, 165, 32},  // gold
	Color{0, 100, 0},     // dark green
	Color{255, 255, 0},   // bright yellow
	Color{0, 0, 255},     // bright blue
	Color{255, 0, 0},     // bright red
	Color{140, 140, 140}, // gray
}

type Screen struct {
	buf      gl.Buffer
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform

	gray   float32
	width  float32
	height float32
}

var triangleData []byte

// Characteristic values of an equilateral triangle in opengl window
// coords.  The base of such a window is two 'units' wide (-1..1), so
// a triangle with side == 2 just fits inside a window.
func computeTriangleLengths() (float32, float32) {
	side := 2.0 / 8.0 // Take a fraction of two, the characteristic size.
	halfBase := side / 2.0
	halfHeight := math.Sqrt(side*side-halfBase*halfBase) / 2.0
	return float32(halfBase), float32(halfHeight)
}

func makeTriangleData() []byte {
	halfBase, halfHeight := computeTriangleLengths()
	return f32.Bytes(binary.LittleEndian,
		-halfBase, -halfHeight, 0.0,
		0.0, halfHeight, 0.0,
		halfBase, -halfHeight, 0.0,
	)
}

func NewScreen() *Screen {
	return &Screen{}
}

func (s *Screen) Start() {
	for i, c := range playerColors {
		playerColors[i] = Color{c.R / 255.0, c.G / 255.0, c.B / 255.0}
	}

	s.buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, s.buf)
	triangleData = makeTriangleData()

	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	var err error
	s.program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("Error in screen.Start: %v", err)
		return
	}
	s.position = gl.GetAttribLocation(s.program, "jrPosition")
	s.color = gl.GetUniformLocation(s.program, "jrColor")
	s.offset = gl.GetUniformLocation(s.program, "jrOffset")
	gl.UseProgram(s.program)
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
	gl.ClearColor(bgRed, bgGreen, bgBlue, opaque)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	s.gray += 0.01
	if s.gray > 1 {
		s.gray = 0
	}

	gl.Uniform4f(s.color, s.gray, 0, 0, opaque)

	gl.EnableVertexAttribArray(s.position)
	gl.VertexAttribPointer(s.position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	for _, b := range balls {
		c := playerColors[b.Owner().Id()%len(playerColors)]
		gl.Uniform4f(s.color, c.R, c.G, c.B, opaque)
		gl.Uniform2f(s.offset, b.GetPos().X/s.width, b.GetPos().Y/s.height)
		gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	}
	gl.DisableVertexAttribArray(s.position)
	// debug.DrawFPS(c)
}

func (s *Screen) Stop() {
	gl.DeleteProgram(s.program)
	gl.DeleteBuffer(s.buf)
}
