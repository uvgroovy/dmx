package main_test

import (
	. "github.com/uvgroovy/dmx/dmxd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uvgroovy/dmx"

	"time"
)

var _ = Describe("Animate", func() {

	It("should get right colors for key frame", func() {
		red := dmx.Color{Red: 0xff}
		green := dmx.Color{Green: 0xff}
		keyframe := KeyFrame{[]dmx.Color{red, green}, time.Second}

		Expect(keyframe.GetColor(0)).To(Equal(red))
		Expect(keyframe.GetColor(1)).To(Equal(green))
		for i := 2; i < 100; i++ {
			Expect(keyframe.GetColor(i)).To(Equal(keyframe.GetColor(1)))
		}
	})

})
