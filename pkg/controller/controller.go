package controller

import (
	"fmt"
	"time"

	"github.com/jhawk7/go-pi-irrigation/pkg/adcsensor"
	"github.com/jhawk7/go-pi-irrigation/pkg/common"
	"github.com/jhawk7/go-pi-irrigation/pkg/pump"
	"periph.io/x/devices/v3/ads1x15"
)

type Controller struct {
	Pump                    *pump.WaterPumpRelay
	ADCSensor               *adcsensor.ADCSensor
	IdealMoisturePercentage float32
	Threshold               int
	Name                    string
	NeedsWater              bool
	LatestReading           float32
	Channel                 ads1x15.Channel
}

func (c *Controller) CheckMoistureLv() {
	reading, readErr := c.ADCSensor.ReadMoistureValue(c.Channel)
	common.ErrorHandler(readErr, false)

	c.LatestReading = reading
	if c.LatestReading <= float32(c.Threshold) {
		c.NeedsWater = true
		common.LogInfo(fmt.Sprintf("%v needs water [reading: %v]", c.Name, c.LatestReading))
		c.PumpWater()
	}

	common.LogInfo(fmt.Sprintf("%v latest reading: %v", c.Name, c.LatestReading))
}

func (c *Controller) PumpWater() {
	for c.LatestReading < float32(c.IdealMoisturePercentage) {
		common.LogInfo(fmt.Sprintf("pumping water for %v", c.Name))
		c.Pump.Release()
		time.Sleep(3 * time.Second)

		reading, readErr := c.ADCSensor.ReadMoistureValue(c.Channel)
		common.ErrorHandler(readErr, true)
		c.LatestReading = reading
	}
	c.NeedsWater = false
}

// This function will be called repeatedly in a go routine
// func (c *Controller) PollMoistureLv() float32 {
// 	latestReading, readErr := c.ADCSensor.ReadMoistureValue(c.Channel)
// 	common.ErrorHandler(readErr, true)
// 	c.LatestReading = latestReading
// 	return latestReading
// }

func (c *Controller) PollMoistureLv() {
	ch := make(chan float32)
	c.ADCSensor.PollMoistureValue(c.Channel, ch)
	for reading := range ch {
		c.LatestReading = reading
	}
}
