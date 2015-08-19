package screen

import (
	"encoding/binary"
	"golang.org/x/mobile/exp/f32"
)

const (
	coordsPerVertex = 3
	vertexCount     = 4

	vertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

	fragmentShader = `#version 100
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
