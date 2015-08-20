package sprite

import (
	"encoding/binary"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
	"math"
)

const (
	coordsPerVertex = 3
	vertexCount     = 4
	opaque          = 1     // 0 is transparent
	anotherBall     = false // testing

	/*
				Coordinates of touch events, ball locations,
				impulses, etc.

				  ( 0, 0) ...  ( W, 0)
				  ...             ...
				  ( 0, H) ...  ( W, H)

				Where W and H are the pixel width and height
				of the screen.

				The {x,y} passed to Draw below is a strictly
				positive normalized screen coordinate, with
				width and height divided out:

				  ( 0, 0) ...  ( 1, 0)
				  ...             ...
				  ( 0, 1) ...  ( 1, 1)

				This must be converted to OpenGL coords:

				  (-1, 1) ...  ( 1, 1)
				  ...    (0, 0)    ...
				  (-1,-1) ...  ( 1,-1)

				This is both a doubling and a sign flip
				for the Y axis.  Transform is:

				      X =  2 * ( oldX - 0.5 ) = 2 oldX - 1
				      Y = -2 * ( oldY - 0.5 ) = 1 - 2 oldY

		    This transform is done by the shader below.
	*/
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

type Ball struct {
	buf      gl.Buffer
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
}

var triangleData []byte

// Characteristic vertex values of an equilateral triangle in opengl
// window coords.  The base of such a window is two 'units' wide
// (-1..1), so a triangle with side == 2 just fits (it's 'huge').
func computeTriangleLengths() (float32, float32) {
	side := 2.0 / 8.0 // A fraction of the characteristic distance.
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

func NewBall() *Ball {
	b := &Ball{}
	b.buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)

	triangleData = makeTriangleData()

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
	gl.UseProgram(b.program)
	return b
}

func (b *Ball) Draw(myColor float32, nx float32, ny float32) {

	gl.Uniform4f(b.color, myColor, 0 /* myColor */, 0, opaque)

	gl.Uniform2f(b.offset, nx, ny)

	gl.EnableVertexAttribArray(b.position)
	gl.VertexAttribPointer(b.position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	gl.DisableVertexAttribArray(b.position)

	if anotherBall {
		gl.Uniform4f(b.color, 0, 0, myColor, opaque)
		gl.Uniform2f(b.offset, nx+.2, ny+.2)

		gl.EnableVertexAttribArray(b.position)
		gl.VertexAttribPointer(b.position, coordsPerVertex, gl.FLOAT, false, 0, 0)
		gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
		gl.DisableVertexAttribArray(b.position)
	}
}

func (b *Ball) Delete() {
	gl.DeleteProgram(b.program)
	gl.DeleteBuffer(b.buf)
}
