package dmx

import (
	"io"
)

type DMXController interface {
	io.Closer
	Write(dmxUniverse *DMXUniverse) error
}

type DMXUniverse struct {
	NumChannels int
	Channels    [512]uint8
}

type Color struct {
	Red, Green, Blue uint8
}

type LightFixture interface {
	SetColor(dmxUniverse *DMXUniverse, c Color)
}
