package k8062

import (
	"fmt"
	"errors"
	"github.com/uvgroovy/go-libusb"
	"github.com/uvgroovy/dmx"
)

const (
	VendorID = 0x10cf // K8062 USB vendor ID
	ProdID   = 0x8062 // K8062 USB product ID
)

func GetDmxControlers() []dmx.DMXController {
	var devices []dmx.DMXController = make([]dmx.DMXController, 0)

	libusb.OpenAllCallback(VendorID, ProdID, func(device *libusb.Device, err error) {
		if err == nil {
			device.Configuration(1)
			device.Interface(0)
			device.Timeout = 200
			devices = append(devices, &K8062DMXController{device: device})
		} else {
			fmt.Errorf("Error opening a device", err)
		}
	})

	return devices
}

type K8062DMXController struct {
	device *libusb.Device
}

func (self K8062DMXController) Close() error {
	self.device.Close()
	return nil
}

func (self K8062DMXController) Write(dmxUniverse *dmx.DMXUniverse) error {
	err := self.sendChannels(dmxUniverse)
	if err != nil {
		return err
	}
	// idk why but it seems that this controller only sets the channels the second time...
	// maybe i have a bug??
	return self.sendChannels(dmxUniverse)
}

func (self K8062DMXController) sendChannels(dmxUniverse *dmx.DMXUniverse) error {
	// copied from dll source code
	var chanIndex = 0

	var numChannel = dmxUniverse.GetNumChannels()
	if numChannel < 8 {
		numChannel = 8
	}

	// remove trailing zeros; will be used later
	// code is weird because it was weird to begin with..
	var advnaceZeros = func() uint8 {
		var n uint8 = 0
		for ;chanIndex < (numChannel - 5); chanIndex++ {
			if dmxUniverse.Channels[chanIndex] != 0 {
				break
			}
			if n == 100 {
				break
			}
			n++
		}
		return n
	}

	var nextValue = func() uint8 {
		var ret = dmxUniverse.Channels[chanIndex]
		chanIndex++
		return ret
	}

	var packet [8]uint8

	packet[0] = 4              // start packet header (4)
	packet[1] = advnaceZeros() // number of zeroes ( not sent ); this saves space on the wire..
	for i := 2; i < len(packet); i++ {
		packet[i] = nextValue() // first ( non-zero ) chan data
	}

	err := self.writeUsb(packet[:])
	if err != nil {
		return err
	}

	// write all the other channels
	for chanIndex <= numChannel {
		// zero packet; probably not needed.
		packet = [8]uint8{}

		if (numChannel - chanIndex) < 6 {
			// not a lot of channels left to fill a packet, send them one by one
			packet[0] = 3 //send one byte of data
			packet[1] = nextValue()

		} else { // (numChannel-chanIndex)>= 6
			// we have 6 or more channels to send; apply compression if possible
			if dmxUniverse.Channels[chanIndex] != 0 {
				packet[0] = 2 // start packet header (2)
				for i := 1; i < len(packet); i++ {
					packet[i] = nextValue()
				}
			} else {
				packet[0] = 5              // start packet header (4)
				packet[1] = advnaceZeros() // number of zeroes ( not sent )
				for i := 2; i < len(packet); i++ {
					packet[i] = nextValue()
				}
			}
		}

		// write the packet to the usb device
		err := self.writeUsb(packet[:])
		if err != nil {
			return err
		}
	}

	return nil
}

func (self K8062DMXController) writeUsb(packet []uint8) error {
	numSent := self.device.InterruptWrite(1, packet)
	if numSent != len(packet) {
		return errors.New("Not all data sent:" + self.device.LastError())
	}
	return nil
}
