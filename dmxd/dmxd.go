package main

import (
	"fmt"
	"github.com/uvgroovy/go-libusb"
	"github.com/uvgroovy/dmx"
	"github.com/uvgroovy/dmx/k8062"
	"time"
)

func sendColors(cont dmx.DMXController, dmxUniverse *dmx.DMXUniverse, devices []dmx.LightFixture, c dmx.Color) {

	for _, d := range devices {
		d.SetColor(dmxUniverse, c)
	}
    
	err := cont.Write(dmxUniverse)
	if err != nil {
		fmt.Println(err)
	}
}

type XL85 struct {
	StartAddress int
}

func (x XL85) SetColor(dmxUniverse *dmx.DMXUniverse, c dmx.Color) {
	base := x.StartAddress
	dmxUniverse.SetChannel(base+0, 100)
	dmxUniverse.SetChannel(base+1, c.Red)
	dmxUniverse.SetChannel(base+2, c.Green)
	dmxUniverse.SetChannel(base+3, c.Blue)
	dmxUniverse.SetChannel(base+4, 250)
	dmxUniverse.SetChannel(base+5, 0)
}

func main() {
	libusb.Init()
	// TOdo check for errors

	dmxControllers := k8062.GetDmxControlers()

	fmt.Printf("Got %d devices\n", len(dmxControllers))

	if len(dmxControllers) == 0 {
		return
	}
	dmxController := dmxControllers[0]

	lightFixtures := make([]dmx.LightFixture, 0)
	for _, d := range dmxControllers {
		defer d.Close()
		/*
			for {
				var err error = errors.New("blah")
				var channel int
				for err != nil {
					fmt.Printf("Channel:")
					_, err = fmt.Scanf("%d", &channel)

				}

				err = errors.New("blah")
				var value int
				for err != nil {
					fmt.Printf("Value:")
					_, err = fmt.Scanf("%d", &value)

				}

				d.Channels[channel] = uint8(value)
				d.SendChannels()

			}*/
		/*
			rand.Seed(time.Now().Unix())
			for i :=0 ;i<512;i++{
				var val uint8 = uint8(rand.Intn(256))
				fmt.Printf("Seeing channel %d to %d\n",i,val)
				d.Channels[i] = val
				d.SendChannels()
				time.Sleep(time.Second)
			}*/
		//	lightFixtures = append(lightFixtures, XL85{1})
		// TODO: read config of config file.
		/*
		 RGBLightFixture{1}
		 RGBLightFixture{4}
		...
		
		*/
		lightFixtures = append(lightFixtures, dmx.RGBLightFixture{1},dmx.RGBLightFixture{1+3*7})
	}

	animate2(dmxController, lightFixtures)
}

func openFixtures() []dmx.LightFixture {
	return nil
}

func animate1(dmxController dmx.DMXController, devices []dmx.LightFixture) {
	var animation = []dmx.Color{dmx.Color{Red: 0xff}, dmx.Color{Blue: 0xff}}
	var wait = 2 * time.Second
	var d dmx.DMXUniverse
	// do whateverzz
	for index := 0; ; index = (index + 1) % len(animation) {
		fmt.Println("send colors")
		sendColors(dmxController, &d, devices, animation[index])
		time.Sleep(wait)
	}

}

func interpolate(c1, c2 dmx.Color, r float32) dmx.Color {
	return dmx.Color{
		uint8(float32(c1.Red)*(1-r) + r*float32(c2.Red)),
		uint8(float32(c1.Green)*(1-r) + r*float32(c2.Green)),
		uint8(float32(c1.Blue)*(1-r) + r*float32(c2.Blue)),
	}
}

func animate2(dmxController dmx.DMXController, devices []dmx.LightFixture) {
	var d dmx.DMXUniverse
	var keyframes = []struct {
		color dmx.Color
		dur   time.Duration
	}{
		{dmx.Color{Red: 0xff}, 2 * time.Second},
		{dmx.Color{Blue: 0xff}, 2 * time.Second},
		{dmx.Color{Green: 0xff}, 2 * time.Second},
		{dmx.Color{Green: 0xff, Blue: 0xff}, 2 * time.Second},
		{dmx.Color{Red: 0xff, Green: 0xff, Blue: 0xff}, 2 * time.Second},
	}

	var frame time.Duration = time.Second / 60

	for index := 0; ; index = (index + 1) % len(keyframes) {
		keyframe1 := keyframes[index]
		keyframe2 := keyframes[(index+1)%len(keyframes)]

		animTime := keyframe1.dur
		ticks := int(animTime / frame)

		for i := 0; i < ticks; i++ {
			var r float32 = float32(i) / float32(ticks)
			newColor := interpolate(keyframe1.color, keyframe2.color, r)
			sendColors(dmxController, &d, devices, newColor)
			time.Sleep(frame)
		}

	}

}
