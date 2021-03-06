package main

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/uvgroovy/dmx"
	"github.com/uvgroovy/dmx/k8062"
	"github.com/uvgroovy/go-libusb"
	"os"
	"time"
)

func sendColors(controllers []dmx.DMXController, devices []dmx.LightFixture, c dmx.Color) {

	var dmxUniverse dmx.DMXUniverse = dmx.DMXUniverse{}
	for _, d := range devices {
		d.SetColor(&dmxUniverse, c)
	}

	for _, cont := range controllers {
		err := cont.Write(&dmxUniverse)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// 100 to support smoother transisions with fast changes
var ShouldAnimate = make(chan bool, 100)
var OscColor dmx.Color

func checkForErrors(dmxControllers []dmx.DMXController) {
	var d dmx.DMXUniverse
	// write zero-universe to check that we can write
	for _, dmxController := range dmxControllers {
		if err := dmxController.Write(&d); err != nil {
			panic(err)
		}
	}
}

func main() {
	libusb.Init()
	// Todo check for errors

	dmxControllers := k8062.GetDmxControlers()
	fmt.Printf("Got %d devices\n", len(dmxControllers))

	checkForErrors(dmxControllers)
	// yuval
	if os.Getenv("OSC_TEST") != "" {
		if len(dmxControllers) == 0 {
			return
		}
	}

	for _, d := range dmxControllers {
		defer d.Close()
	}
	gui := true
	if gui {
		dmxControllers = append(dmxControllers, <-GetGuiController())
		gdk.ThreadsInit()
	}

	lightFixtures := openFixtures()

	keyframes := getKeyFrames()

	setupOsc()

	worker(keyframes, dmxControllers, lightFixtures)

}

func GetGuiController() <-chan *GuiController {
	controler := make(chan *GuiController, 1)

	go func() {
		gtk.Init(nil)
		controler <- CreateGuiController()
		gtk.Main()
	}()

	return controler
}

// implements DMXController
type GuiController struct {
	buttons []*gtk.Button
}

func CreateGuiController() *GuiController {
	guiController := &GuiController{}
	guiController.buttons = make([]*gtk.Button, 0)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("GTK Go!")
	window.SetIconName("gtk-dialog-info")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		fmt.Println("got destroy!", ctx.Data().(string))
		gtk.MainQuit()
	}, "foo")

	buttonsBox := gtk.NewHBox(false, 1)

	black := gdk.NewColorRGB(0, 0, 0)

	for i := 0; i < 8; i++ {
		button := gtk.NewButtonWithLabel(fmt.Sprint(i))

		button.ModifyBG(gtk.STATE_NORMAL, black)
		guiController.buttons = append(guiController.buttons, button)
		buttonsBox.Add(button)
	}

	window.Add(buttonsBox)
	window.SetSizeRequest(600, 600)
	window.ShowAll()

	return guiController
}

func (g *GuiController) Close() error {
	return nil
}
func (g *GuiController) Write(dmxUniverse *dmx.DMXUniverse) error {
	gdk.ThreadsEnter()
	for i := 0; i < 3*8; i += 3 {

		c := gdk.NewColorRGB(dmxUniverse.Channels[i], dmxUniverse.Channels[i+1], dmxUniverse.Channels[i+2])

		g.buttons[i/3].ModifyBG(gtk.STATE_NORMAL, c)
	}

	gdk.ThreadsLeave()

	return nil
}

func shouldAnimate() bool {

	// grab a valid value from the channel (wait if needed)
	var shouldAnimate bool = <-ShouldAnimate
	// if channel still not empty clear it out and take the last value
	for {
		select {
		case shouldAnimate = <-ShouldAnimate:
		default:
			return shouldAnimate
		}
	}
}

func worker(keyframes KeyFrames, dmxControllers []dmx.DMXController, lightFixtures []dmx.LightFixture) {
	// request to stop animation signal
	stop := make(chan bool, 1)
	// stopping complete signal
	stopped := make(chan bool, 1)

	// we start with no animation (obviously as we just started!)
	currentlyAnimating := false

	// we want to animate by default
	ShouldAnimate <- true

	// this will work forever
	for {
		if shouldAnimate() {
			fmt.Println("should animate")
			if !currentlyAnimating {
				go func() {
					fmt.Println("Anmation start")
					keyframes.Animate(dmxControllers, lightFixtures, stop)
					fmt.Println("Anmation end")
					stopped <- true
				}()
				currentlyAnimating = true
			}
		} else {
			fmt.Println("should not animate")

			if currentlyAnimating {

				fmt.Println("waiting for animation to stop")
				stop <- true
				<-stopped
			}
			currentlyAnimating = false
			fmt.Println("animation stopped")

			sendColors(dmxControllers, lightFixtures, OscColor)

		}

	}

}

func getBoolFromMessage(msg *osc.Message) (bool, bool) {

	if len(msg.Arguments) != 1 {
		return false, false
	}
	value := msg.Arguments[0]
	switch value.(type) {
	case int32:
		return 0 != (value.(int32)), true
	case float32:
		return 0 != (value.(float32)), true
	}

	return false, false
}

func getColorFromMessage(msg *osc.Message) (uint8, bool) {
	if len(msg.Arguments) != 1 {
		return 0, false
	}

	color := msg.Arguments[0]
	switch color.(type) {
	case float32:
		return uint8(color.(float32)), true
	case float64:
		return uint8(color.(float64)), true
	case int32:
		return uint8(color.(int32)), true
	}

	return 0, false
}

func setupOsc() {

	addr := "0.0.0.0:12000"
	server := &osc.Server{Addr: addr}
	var shouldAnimate = true

	server.Handle("/red", func(msg *osc.Message) {
		if red, ok := getColorFromMessage(msg); ok && red != OscColor.Red {
			fmt.Println("Got red", red)
			OscColor.Red = uint8(red)
			ShouldAnimate <- shouldAnimate
		}
	})
	server.Handle("/green", func(msg *osc.Message) {
		if green, ok := getColorFromMessage(msg); ok && green != OscColor.Green {
			fmt.Println("Got green", green)
			OscColor.Green = uint8(green)
			ShouldAnimate <- shouldAnimate
		}
	})
	server.Handle("/blue", func(msg *osc.Message) {
		if blue, ok := getColorFromMessage(msg); ok && blue != OscColor.Blue {
			fmt.Println("Got blue", blue)
			OscColor.Blue = uint8(blue)
			ShouldAnimate <- shouldAnimate
		}
	})
	server.Handle("/button", func(msg *osc.Message) {

		if value, ok := getBoolFromMessage(msg); ok {
			shouldAnimate = value
			fmt.Println("Got button", value)
		} else {
			shouldAnimate = !shouldAnimate
		}

		fmt.Println("Got button")
		ShouldAnimate <- shouldAnimate
	})

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
}

func openFixtures() []dmx.LightFixture {

	lightFixtures := make([]dmx.LightFixture, 0)

	for i := 0; i < 8; i++ {
		lightFixtures = append(lightFixtures, dmx.RGBLightFixture{1 + i*3})
	}
	return lightFixtures
}

var DefaultAnimation KeyFrames = []KeyFrame{
	{[]dmx.Color{dmx.Color{Red: 0xff}}, time.Second},
	{[]dmx.Color{dmx.Color{Blue: 0xff}}, 2 * time.Second},
	{[]dmx.Color{dmx.Color{Green: 0xff}}, 2 * time.Second},
	{[]dmx.Color{dmx.Color{Green: 0xff, Blue: 0xff}}, 2 * time.Second},
	{[]dmx.Color{dmx.Color{Red: 0xff, Green: 0x11, Blue: 0x11}}, 2 * time.Second},
}

func getKeyFrames() KeyFrames {
	var input *os.File

	if len(os.Args) == 2 {
		if os.Args[1] == "-" {
			input = os.Stdin
			fmt.Println("Reading animation from stdin")
		} else {
			var err error
			input, err = os.Open(os.Args[1])
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

func animate1(dmxControllers []dmx.DMXController, devices []dmx.LightFixture) {
	var animation = []dmx.Color{dmx.Color{Red: 0xff}, dmx.Color{Blue: 0xff}}
	var wait = 2 * time.Second
	// do whateverzz
	for index := 0; ; index = (index + 1) % len(animation) {
		fmt.Println("send colors")
		sendColors(dmxControllers, devices, animation[index])
		time.Sleep(wait)
	}

}
