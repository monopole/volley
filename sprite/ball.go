package sprite

import (
	"encoding/binary"
	"github.com/monopole/croupier/model"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
	"math"
	"math/rand"
)

const (
	extraBalls      = 10
	coordsPerVertex = 3
	vertexCount     = 4
	opaque          = 1    // 0 is transparent
	anotherBall     = true // testing

	/*

		Coordinates of touch events, ball locations,
		impulses, etc. are in screen coords

		  ( 0, 0) ...  ( W, 0)
		  ...             ...
		  ( 0, H) ...  ( W, H)

		where W and H are the pixel width and height
		of the window.

		The {x,y} passed to Draw below is assumed to
		be a non-negative normalized screen coordinate, with
		width and height divided out:

		  ( 0, 0) ...  ( 1, 0)
		  ...             ...
		  ( 0, 1) ...  ( 1, 1)

		This must be converted to OpenGL coords:

		  (-1, 1) ...  ( 1, 1)
		  ...    (0, 0)    ...
		  (-1,-1) ...  ( 1,-1)

		This is both a doubling and a sign flip
		for the Y axis.  The transform is:

		  X =  2 * ( oldX - 0.5 ) = 2 oldX - 1
		  Y = -2 * ( oldY - 0.5 ) = 1 - 2 oldY

		and is performed by the shader below.  It could
		be done in Go on the CPU, but might as well let
		the GPU contribute.

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
	buf       gl.Buffer
	program   gl.Program
	position  gl.Attrib
	offset    gl.Uniform
	color     gl.Uniform
	moarBalls []model.Vec
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

func NewBall() *Ball {

	b := &Ball{}

	b.moarBalls = make([]model.Vec, extraBalls)
	for i := 0; i < extraBalls; i++ {
		b.moarBalls[i] = model.Vec{rand.Float32(), rand.Float32()}
	}

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

	gl.Uniform4f(b.color, 0, 1.0, 1.0, opaque)
	if len(b.moarBalls) > 0 {
		for _, v := range b.moarBalls {
			gl.Uniform2f(b.offset, v.X, v.Y)
			gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
		}
	}
	gl.DisableVertexAttribArray(b.position)

}

func (b *Ball) Delete() {
	gl.DeleteProgram(b.program)
	gl.DeleteBuffer(b.buf)
}
