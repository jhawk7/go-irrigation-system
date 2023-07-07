package controller

import (
	"fmt"
	"time"

	"github.com/jhawk7/go-pi-irrigation/pkg/common"
	"github.com/jhawk7/go-pi-irrigation/pkg/moisture_sensor"
	"github.com/jhawk7/go-pi-irrigation/pkg/water_pump"
	"periph.io/x/devices/v3/ads1x15"
)

type Controller struct {
	Pump                    *water_pump.WaterPumpRelay
	Msensor                 *moisture_sensor.ADCMoistureSensor
	IdealMoisturePercentage int
	Threshold               int
	Name                    string
	NeedsWater              bool
	LatestReading           float32
	Channel                 ads1x15.Channel
}

func (c *Controller) CheckMoistureLv() {
	reading, readErr := c.Msensor.ReadMoistureValue(c.Channel)
	common.ErrorHandler(readErr, false)

	c.LatestReading = reading
	if c.LatestReading <= float32(c.Threshold) {
		c.NeedsWater = true
		common.LogInfo(fmt.Sprintf("%v needs water; [reading: %v]", c.Name, c.LatestReading))
	}

	for c.NeedsWater {
		c.PumpWater()
		latestReading, readErr := c.Msensor.ReadMoistureValue(c.Channel)
		common.ErrorHandler(readErr, true)
		if latestReading >= float32(c.IdealMoisturePercentage) {
			c.LatestReading = latestReading
			c.NeedsWater = false
		}
	}

	common.LogInfo(fmt.Sprintf("%v latest reading: %v", c.Name, c.LatestReading))
}

// This function will be called repeatedly in a go routine
// func (c *Controller) PollMoistureLv() float32 {
// 	latestReading, readErr := c.Msensor.ReadMoistureValue(c.Channel)
// 	common.ErrorHandler(readErr, true)
// 	c.LatestReading = latestReading
// 	return latestReading
// }

func (c *Controller) PollMoistureLv() {
	ch := make(chan float32)
	c.Msensor.PollMoistureValue(c.Channel, ch)
	for reading := range ch {
		c.LatestReading = reading
	}
}

func (c *Controller) PumpWater() {
	common.LogInfo(fmt.Sprintf("pumping water for %v", c.Name))
	c.Pump.Release()
	time.Sleep(3 * time.Second)
}
