package main

import (
	"os"
	"encoding/hex"
	"encoding/json"
	"io"
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
	for _, d := range dmxControllers {
		defer d.Close()
	}
	
	dmxController := dmxControllers[0]

	lightFixtures := openFixtures()
	
	animate2(getKeyFrames(), dmxController, lightFixtures)
	
	
}

func openFixtures() []dmx.LightFixture {

	lightFixtures := make([]dmx.LightFixture, 0)
	
	for i := 0; i < 12; i++ {
		lightFixtures = append(lightFixtures, 	dmx.RGBLightFixture{1 + i*3})
	}
	return lightFixtures
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

func white(dmxController dmx.DMXController, devices []dmx.LightFixture) {
	var d dmx.DMXUniverse
	// do whateverzz
	sendColors(dmxController, &d, devices, dmx.Color{Red: 0xff})
	sendColors(dmxController, &d, devices, dmx.Color{Red: 0xff})
	
	time.Sleep(time.Second*2)
	sendColors(dmxController, &d, devices, dmx.Color{Red: 0xff, Green: 0xff, Blue: 0xff})

}

func interpolate(c1, c2 dmx.Color, r float32) dmx.Color {
	return dmx.Color{
		uint8(float32(c1.Red)*(1-r) + r*float32(c2.Red)),
		uint8(float32(c1.Green)*(1-r) + r*float32(c2.Green)),
		uint8(float32(c1.Blue)*(1-r) + r*float32(c2.Blue)),
	}
}

type Keyframe struct {
		Color dmx.Color
		Duration   time.Duration
}

func parseColor(color string) dmx.Color {
	var data [1]byte
	hex.Decode(data[:], []byte(color[:2]))
	r := data[0]
	hex.Decode(data[:], []byte(color[2:4]))
	g := data[0]
	hex.Decode(data[:], []byte(color[4:5]))
	b := data[0]
	return dmx.Color{r,g,b};
}

func ReadKeyframes(reader io.Reader) []Keyframe {
	keyframes := make([]Keyframe,0)
	type KeyframeJson struct {
		Color string
		Duration  int
	}
		
	dec := json.NewDecoder(reader)
	for {
		var frame KeyframeJson
		if err := dec.Decode(&frame); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		curDuration := int64(time.Millisecond) * int64(frame.Duration)
		curKeyFrame := Keyframe{Color: parseColor(frame.Color), Duration: time.Duration(curDuration)}
		keyframes = append(keyframes, curKeyFrame)
		fmt.Printf("%v\n", curKeyFrame)
	}
	
	return keyframes
}

var DefaultAnimation = []Keyframe{
	{dmx.Color{Red: 0xff},  time.Second},
	{dmx.Color{Blue: 0xff}, 2 * time.Second},
	{dmx.Color{Green: 0xff}, 2 * time.Second},
	{dmx.Color{Green: 0xff, Blue: 0xff}, 2 * time.Second},
	{dmx.Color{Red: 0xff, Green: 0xff, Blue: 0xff}, 2 * time.Second},
}

func getKeyFrames() []Keyframe {
	var input *os.File
	
	if len(os.Args) == 2 {
		if os.Args[1] == "-" {
			input = os.Stdin	
			fmt.Println("Reading animation from stdin")
		} else {
			input,err := os.Open(os.Args[1])
			if err != nil {
				panic("can't open animation file")
			}
			defer input.Close()	
			fmt.Println("Reading animation from file", os.Args[1])
		}
		
		return ReadKeyframes(input)
	} else {
		fmt.Println("Using default animation")
		return DefaultAnimation
	}
}

func animate2(animation []Keyframe, dmxController dmx.DMXController, devices []dmx.LightFixture) {
	var d dmx.DMXUniverse

	var frame time.Duration = time.Second / 30

	for index := 0; ; index = (index + 1) % len(animation) {
		keyframe1 := animation[index]
		keyframe2 := animation[(index+1)%len(animation)]

		animTime := keyframe1.Duration
		ticks := int(animTime / frame)

		for i := 0; i < ticks; i++ {
			var r float32 = float32(i) / float32(ticks)
			newColor := interpolate(keyframe1.Color, keyframe2.Color, r)
			sendColors(dmxController, &d, devices, newColor)
			time.Sleep(frame)
		}

	}

}
