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

func (dmxUniverse *DMXUniverse) SetChannel(address  int, value uint8) {
	currentChannels := address + 1
	if dmxUniverse.NumChannels < currentChannels {
      dmxUniverse.NumChannels = currentChannels
   	}
	dmxUniverse.Channels[address] = value
}

type Color struct {
	Red, Green, Blue uint8
}

type LightFixture interface {
	SetColor(dmxUniverse *DMXUniverse, c Color)
}

type RGBLightFixture struct {
	StartAddress int
}

func (x RGBLightFixture) SetColor(dmxUniverse *DMXUniverse, c Color) {
	dmxUniverse.SetChannel(x.StartAddress  , c.Red)
	dmxUniverse.SetChannel(x.StartAddress+1, c.Green)
	dmxUniverse.SetChannel(x.StartAddress+2, c.Blue)
}
