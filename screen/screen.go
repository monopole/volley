package screen

import (
	"encoding/binary"
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
)

var (
	program      gl.Program
	position     gl.Attrib
	offset       gl.Uniform
	color        gl.Uniform
	buf          gl.Buffer
	green        float32
	red          float32
	blue         float32
	gray         float32
	touchX       float32
	touchY       float32
	iHaveTheCard bool
)

type Screen struct {
}

func NewScreen() *Screen {
	gray = 0.1
	red = 0.4
	green = 0.4
	blue = 0.4

	return &Screen{}
}

func (s *Screen) onPaint(c config.Event) {
	if iHaveTheCard {
		red = 0.8
	} else {
		red = 0.4
	}
	gl.ClearColor(red, green, blue, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(program)

	gray += 0.01
	if gray > 1 {
		gray = 0
	}
	// Color the triangle
	gl.Uniform4f(color, gray, 0, gray, 1)
	// Move the triangle
	gl.Uniform2f(offset, touchX/float32(c.WidthPx), touchY/float32(c.HeightPx))

	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.EnableVertexAttribArray(position)
	gl.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	if iHaveTheCard {
		gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	}
	gl.DisableVertexAttribArray(position)

	debug.DrawFPS(c)
}

func (s *Screen) onStart() {
	var err error
	program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	position = gl.GetAttribLocation(program, "position")
	color = gl.GetUniformLocation(program, "color")
	offset = gl.GetUniformLocation(program, "offset")

	// TODO(crawshaw): the debug package needs to put GL state init here
	// Can this be an app.RegisterFilter call now??
}

func (s *Screen) onStop() {
	gl.DeleteProgram(program)
	gl.DeleteBuffer(buf)
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
