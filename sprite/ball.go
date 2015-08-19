package sprite

import (
	"encoding/binary"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
)

const (
	coordsPerVertex = 3
	vertexCount     = 4
	opaque          = 1 // 0 is transparent

	vertexShader = `#version 100
uniform vec2 jrOffset;
attribute vec4 jrPosition;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
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

var triangleData = f32.Bytes(binary.LittleEndian,
	// x, y, z, in percentage distance from origin to window border
	// counter clockwise.
	-0.1, 0.2, 0.0,
	0.1, 0.2, 0.0,
	0.1, -0.2, 0.0,
)

type Ball struct {
	buf      gl.Buffer
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
}

func NewBall() *Ball {
	b := &Ball{}
	b.buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)
	var err error
	b.program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return nil
	}
	b.position = gl.GetAttribLocation(b.program, "jrPosition")
	b.color = gl.GetUniformLocation(b.program, "jrColor")
	b.offset = gl.GetUniformLocation(b.program, "jrOffset")
	return b
}

func (b *Ball) Draw(myColor float32, nx float32, ny float32) {
	gl.UseProgram(b.program)

	gl.Uniform4f(b.color, myColor, 0, myColor, opaque)
	gl.Uniform2f(b.offset, nx, ny)

	gl.EnableVertexAttribArray(b.position)
	gl.VertexAttribPointer(b.position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	gl.DisableVertexAttribArray(b.position)
}

func (b *Ball) Delete() {
	gl.DeleteProgram(b.program)
	gl.DeleteBuffer(b.buf)
}
