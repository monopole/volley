package sprite

import (
	"encoding/binary"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

const (
	CoordsPerVertex = 3
	VertexCount     = 4

	VertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

	FragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
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
	buf gl.Buffer
}

func NewBall() *Ball {
	b := &Ball{}
	b.buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)
	return b
}

func (b *Ball) Buffer() gl.Buffer {
	return b.buf
}

func (b *Ball) DeleteBuffer() {
	gl.DeleteBuffer(b.buf)
}
