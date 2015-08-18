package dmx

import (
	"io"
)

type DMXController interface {
	io.Closer
	Write(dmxUniverse *DMXUniverse) error
}

type DMXUniverse struct {
	Channels    [512]uint8
}

func (dmxUniverse *DMXUniverse) SetChannel(address  int, value uint8) {
	dmxUniverse.Channels[address] = value
}

func (dmxUniverse *DMXUniverse) GetNumChannels() int {
	for i := len(dmxUniverse.Channels); i > 0; i-- {
		if (dmxUniverse.Channels[i-1] != 0) {
			return i
		}
	}
	return 0
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
