package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/uvgroovy/dmx"
	"io"
	"time"
)

func interpolate(c1, c2 dmx.Color, r float32) dmx.Color {
	return dmx.Color{
		uint8(float32(c1.Red)*(1-r) + r*float32(c2.Red)),
		uint8(float32(c1.Green)*(1-r) + r*float32(c2.Green)),
		uint8(float32(c1.Blue)*(1-r) + r*float32(c2.Blue)),
	}
}

type KeyFrame struct {
	Colors   []dmx.Color
	Duration time.Duration
}

func parseColor(colors []string) []dmx.Color {
	dmxcolors := make([]dmx.Color, 0, len(colors))
	for _, color := range colors {
		var data [1]byte
		_, err := hex.Decode(data[:], []byte(color[:2]))
		if err != nil {
			panic(err)
		}
		r := data[0]
		_, err = hex.Decode(data[:], []byte(color[2:4]))
		if err != nil {
			panic(err)
		}
		g := data[0]
		_, err = hex.Decode(data[:], []byte(color[4:6]))
		if err != nil {
			panic(err)
		}
		b := data[0]

		dmxcolors = append(dmxcolors, dmx.Color{r, g, b})
	}

	return dmxcolors
}

func ReadKeyframes(reader io.Reader) KeyFrames {
	keyframes := make([]KeyFrame, 0)
	type KeyframeJson struct {
		Colors   []string
		Duration float64
	}
	dec := json.NewDecoder(reader)
	for {
		var frame KeyframeJson
		if err := dec.Decode(&frame); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		curDuration := int64(time.Millisecond) * int64(frame.Duration*1000)
		curKeyFrame := KeyFrame{Colors: parseColor(frame.Colors), Duration: time.Duration(curDuration)}
		keyframes = append(keyframes, curKeyFrame)
		fmt.Println("Current key frame", curKeyFrame)
	}

	return keyframes
}

func getColor(keyframe KeyFrame, index int) dmx.Color {
	if index < len(keyframe.Colors) {
		return keyframe.Colors[index]
	}

	return keyframe.Colors[len(keyframe.Colors)-1]

}

type KeyFrames []KeyFrame

func (animation KeyFrames) Animate(dmxControllers []dmx.DMXController, fixtures []dmx.LightFixture, stop <-chan bool) {
	var frame time.Duration = time.Second / 30
	var dmxUniverse dmx.DMXUniverse = dmx.DMXUniverse{}

	for index := 0; ; index = (index + 1) % len(animation) {
		keyframe1 := animation[index]
		keyframe2 := animation[(index+1)%len(animation)]

		animTime := keyframe1.Duration
		ticks := int(animTime / frame)

		for i := 0; i < ticks; i++ {
			var r float32 = float32(i) / float32(ticks)
			for index, fixture := range fixtures {
				newColor := interpolate(getColor(keyframe1, index), getColor(keyframe2, index), r)
				fixture.SetColor(&dmxUniverse, newColor)
			}

			for _, dmxController := range dmxControllers {
				err := dmxController.Write(&dmxUniverse)
				if err != nil {
					fmt.Println(err)
				}
			}

			select {
			case <-stop:
				return
			case <-time.After(frame):

			}
		}

	}

}
