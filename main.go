// +build darwin linux

// Hacked up version of
// https://godoc.org/golang.org/x/mobile/example/basic

package main

import (
	"encoding/binary"
	"github.com/monopole/croupier/game"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
	"v.io/v23"
	"v.io/x/ref/lib/signals"
	_ "v.io/x/ref/runtime/factories/generic"
)

var (
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	buf      gl.Buffer

	gray float32

	green float32
	red   float32
	blue  float32

	touchX float32
	touchY float32

	iHaveTheCard bool
)

func main() {
	app.Main(func(a app.App) {

		ctx, shutdown := v23.Init()
		defer shutdown()

		gm := game.NewGameManager(ctx)

		// The server initializes with player '0' holding the card.
		iHaveTheCard = gm.MyNumber() == 0

		log.Printf("Hi there.\n")

		red = 0.4
		green = 0.4
		blue = 0.4

		var c config.Event
		pollCounter := 0
		for e := range a.Events() {
			pollCounter++
			if pollCounter == 30 { // 60 ~= one second
				iHaveTheCard = gm.MyNumber() == gm.WhoHasTheCard()
				pollCounter = 0
			}
			switch e := app.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					onStart()
				case lifecycle.CrossOff:
					onStop()
				}
			case config.Event:
				c = e
				touchX = float32(c.WidthPx / 2)
				touchY = float32(c.HeightPx / 2)
				gm.SetOrigin(touchX, touchY)
			case paint.Event:
				onPaint(c)
				a.EndPaint(e)
			case touch.Event:
				if e.Type == touch.TypeEnd && iHaveTheCard {
					gm.PassTheCard()
					touchX = gm.GetOriginX()
					touchY = gm.GetOriginY()
					iHaveTheCard = false
				} else {
					touchX = e.X
					touchY = e.Y
				}
			}

		}
		// Normal means to end v23 services:
		<-signals.ShutdownOnSignals(ctx)

	})
}

func onStart() {
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

func onStop() {
	gl.DeleteProgram(program)
	gl.DeleteBuffer(buf)
}

func onPaint(c config.Event) {
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
